/*
 * Copyright (c) 2024, Psiphon Inc.
 * All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package common

import (
	"container/list"
	"fmt"
	"sync"

	tls "github.com/Psiphon-Labs/psiphon-tls"
	utls "github.com/Psiphon-Labs/utls"
)

// LRUClientSessionCache is a ClientSessionCache implementation that uses an LRU
// caching strategy.
type LRUClientSessionCache struct {
	sync.Mutex

	m        map[string]*list.Element
	q        *list.List
	capacity int
}

type lruSessionCacheEntry struct {
	sessionKey string
	ticket     []byte
	state      []byte
}

// NewLRUClientSessionCache returns a [ClientSessionCache] with the given
// capacity that uses an LRU strategy. If capacity is < 1, a default capacity
// is used instead.
func NewLRUClientSessionCache(capacity int) *LRUClientSessionCache {
	const defaultSessionCacheCapacity = 64

	if capacity < 1 {
		capacity = defaultSessionCacheCapacity
	}
	return &LRUClientSessionCache{
		m:        make(map[string]*list.Element),
		q:        list.New(),
		capacity: capacity,
	}
}

// Put adds the provided (sessionKey, cs) pair to the cache. If cs is nil, the entry
// corresponding to sessionKey is removed from the cache instead.
func (c *LRUClientSessionCache) Put(sessionKey string, ticket, state []byte) {
	c.Lock()
	defer c.Unlock()

	if elem, ok := c.m[sessionKey]; ok {
		if state == nil {
			c.q.Remove(elem)
			delete(c.m, sessionKey)

			fmt.Printf("[LRUCache] Removed %s\n", sessionKey)

		} else {
			entry := elem.Value.(*lruSessionCacheEntry)
			entry.ticket = ticket
			entry.state = state
			c.q.MoveToFront(elem)

			fmt.Printf("[LRUCache] Updated %s\n", sessionKey)

		}
		return
	}
	if c.q.Len() < c.capacity {
		entry := &lruSessionCacheEntry{sessionKey, ticket, state}
		c.m[sessionKey] = c.q.PushFront(entry)

		fmt.Printf("[LRUCache] Pushed %s\n", sessionKey)

		return
	}

	elem := c.q.Back()
	entry := elem.Value.(*lruSessionCacheEntry)
	delete(c.m, entry.sessionKey)
	entry.sessionKey = sessionKey
	entry.state = state
	entry.ticket = ticket
	c.q.MoveToFront(elem)
	c.m[sessionKey] = elem

	fmt.Printf("[LRUCache] Evicted and replaced %s\n", entry.sessionKey)
}

// Get returns the [ClientSessionState] value associated with a given key. It
// returns (nil, false) if no value is found.
func (c *LRUClientSessionCache) Get(sessionKey string) (ticket, state []byte, ok bool) {
	c.Lock()
	defer c.Unlock()

	if elem, ok := c.m[sessionKey]; ok {

		fmt.Printf("[LRUCache] Found %s\n", sessionKey)

		c.q.MoveToFront(elem)
		entry := elem.Value.(*lruSessionCacheEntry)
		return entry.ticket, entry.state, true
	}

	fmt.Printf("[LRUCache] Not found %s\n", sessionKey)

	return nil, nil, false
}

const TLS_NULL_SESSION_KEY = ""

type ClientSessionStateType interface {
	*tls.ClientSessionState | *utls.ClientSessionState
}

// SessionCache is a generic interface constraint for session cache types.
// It abstracts the Get and Put methods for different session state types.
type SessionCache[S ClientSessionStateType] interface {
	Get(sessionKey string) (session S, ok bool)
	Put(sessionKey string, session S)
}

type CacheWrapper[S ClientSessionStateType, C SessionCache[S]] struct {
	cache *LRUClientSessionCache
}

func WrapLRUCache[S ClientSessionStateType, C SessionCache[S]](cache *LRUClientSessionCache) SessionCache[S] {
	return &CacheWrapper[S, C]{cache}
}

func (c *CacheWrapper[S, C]) Get(sessionKey string) (S, bool) {

	ticket, state, ok := c.cache.Get(sessionKey)
	if !ok {
		var zeroS S
		return zeroS, false
	}

	var instanceS S
	switch any(instanceS).(type) {
	case *tls.ClientSessionState:
		ss, err := tls.ParseSessionState(state)
		if err != nil {
			return nil, false
		}
		cs, err := tls.NewResumptionState(ticket, ss)
		if err != nil {
			return nil, false
		}

		// cast cs to S
		return any(cs).(S), true

	case *utls.ClientSessionState:
		ss, err := utls.ParseSessionState(state)
		if err != nil {
			return nil, false
		}
		cs, err := utls.NewResumptionState(ticket, ss)
		if err != nil {
			return nil, false
		}
		return any(cs).(S), true
	default:
		panic(fmt.Sprintf("Unsupported session state type: %T", instanceS))
	}
}

func (c *CacheWrapper[S, C]) Put(sessionKey string, cs S) {
	switch cs := any(cs).(type) {
	case *tls.ClientSessionState:
		ticket, state, err := cs.ResumptionState()
		if state == nil || err != nil {
			return
		}
		stateBytes, err := state.Bytes()
		if err != nil {
			return
		}
		c.cache.Put(sessionKey, ticket, stateBytes)
	case *utls.ClientSessionState:
		ticket, state, err := cs.ResumptionState()
		if state == nil || err != nil {
			return
		}
		stateBytes, err := state.Bytes()
		if err != nil {
			return
		}
		c.cache.Put(sessionKey, ticket, stateBytes)
	default:
		panic(fmt.Sprintf("Unsupported session state type: %T", cs))
	}
}

// HardKVCacheWrapper is a generic wrapper around a ClientSessionCache
// that provides a hard-coded key for the cache.
// S is the session state type (e.g., *tls.ClientSessionState).
// C is the cache type (e.g., tls.ClientSessionCache) implementing SessionCache[S].
type HardKVCacheWrapper[S ClientSessionStateType, C SessionCache[S]] struct {
	ClientSessionCache CacheWrapper[S, C]
	sessionKey         string
}

// NewHardKVCacheWrapper creates a new HardKVCacheWrapper.
// If the hardCodedSessionKey is empty (TLS_NULL_SESSION_KEY), SetSessionKey has to be called
// before using the cache.
func NewHardKVCacheWrapper[S ClientSessionStateType, C SessionCache[S]](
	cache *LRUClientSessionCache,
	hardCodedSessionKey string,
) *HardKVCacheWrapper[S, C] {
	return &HardKVCacheWrapper[S, C]{
		ClientSessionCache: CacheWrapper[S, C]{cache},
		sessionKey:         hardCodedSessionKey,
	}
}

// Get retrieves the session from the cache using the hard-coded session key.
// The key argument is ignored.
func (c *HardKVCacheWrapper[S, C]) Get(_ string) (session S, ok bool) {
	if c.sessionKey == TLS_NULL_SESSION_KEY {
		var zeroS S
		return zeroS, false
	}
	return c.ClientSessionCache.Get(c.sessionKey)
}

// Put stores the session in the cache using the hard-coded session key.
// The key argument is ignored.
func (c *HardKVCacheWrapper[S, C]) Put(_ string, cs S) {
	if c.sessionKey == TLS_NULL_SESSION_KEY {
		return
	}
	c.ClientSessionCache.Put(c.sessionKey, cs)
}

// RemoveCacheEntry removes the cache entry for the hard-coded session key.
func (c *HardKVCacheWrapper[S, C]) RemoveCacheEntry() {
	if c.sessionKey == TLS_NULL_SESSION_KEY {
		return
	}
	var nilS S // Creates a nil value for pointer/interface types, or zero value otherwise
	c.ClientSessionCache.Put(c.sessionKey, nilS)
}

// SetSessionKey sets the hard-coded session key if not already set.
func (c *HardKVCacheWrapper[S, C]) SetSessionKey(key string) {
	if c.sessionKey != TLS_NULL_SESSION_KEY {
		return
	}
	c.sessionKey = key
}

// TLSClientSessionCacheWrapper is a type alias for HardKVCacheWrapper specialized
// for tls.ClientSessionCache.
type TLSClientSessionCacheWrapper = HardKVCacheWrapper[*tls.ClientSessionState, tls.ClientSessionCache]

func WrapTLSClientSessionCache(cache *LRUClientSessionCache, sessionKey string) *HardKVCacheWrapper[*tls.ClientSessionState, tls.ClientSessionCache] {
	return &HardKVCacheWrapper[*tls.ClientSessionState, tls.ClientSessionCache]{
		ClientSessionCache: CacheWrapper[*tls.ClientSessionState, tls.ClientSessionCache]{cache},
		sessionKey:         sessionKey,
	}
}

// for utls.ClientSessionCache.
type UTLSClientSessionCacheWrapper = HardKVCacheWrapper[*utls.ClientSessionState, utls.ClientSessionCache]

func WrapUTLSClientSessionCache(cache *LRUClientSessionCache, sessionKey string) *HardKVCacheWrapper[*utls.ClientSessionState, utls.ClientSessionCache] {
	return &HardKVCacheWrapper[*utls.ClientSessionState, utls.ClientSessionCache]{
		ClientSessionCache: CacheWrapper[*utls.ClientSessionState, utls.ClientSessionCache]{cache},
		sessionKey:         sessionKey,
	}
}
