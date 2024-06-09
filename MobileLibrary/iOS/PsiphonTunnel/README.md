### Note on Apple platform targets.

Apple defined conditionals, from `TargetConditionals.h`.


```
/*
 *  TARGET_OS_*
 *
 *  These conditionals specify in which Operating System the generated code will
 *  run.  Indention is used to show which conditionals are evolutionary subclasses.
 *
 *  The MAC/WIN32/UNIX conditionals are mutually exclusive.
 *  The IOS/TV/WATCH/VISION conditionals are mutually exclusive.
 *
 *    TARGET_OS_WIN32              - Generated code will run on WIN32 API
 *    TARGET_OS_WINDOWS            - Generated code will run on Windows
 *    TARGET_OS_UNIX               - Generated code will run on some Unix (not macOS)
 *    TARGET_OS_LINUX              - Generated code will run on Linux
 *    TARGET_OS_MAC                - Generated code will run on a variant of macOS
 *      TARGET_OS_OSX                - Generated code will run on macOS
 *      TARGET_OS_IPHONE             - Generated code will run on a variant of iOS (firmware, devices, simulator)
 *        TARGET_OS_IOS                - Generated code will run on iOS
 *          TARGET_OS_MACCATALYST        - Generated code will run on macOS
 *        TARGET_OS_TV                 - Generated code will run on tvOS
 *        TARGET_OS_WATCH              - Generated code will run on watchOS
 *        TARGET_OS_VISION             - Generated code will run on visionOS
 *        TARGET_OS_BRIDGE             - Generated code will run on bridge devices
 *      TARGET_OS_SIMULATOR          - Generated code will run on an iOS, tvOS, watchOS, or visionOS simulator
 *      TARGET_OS_DRIVERKIT          - Generated code will run on macOS, iOS, tvOS, watchOS, or visionOS
 *
 *    TARGET_OS_EMBEDDED           - DEPRECATED: Use TARGET_OS_IPHONE and/or TARGET_OS_SIMULATOR instead
 *    TARGET_IPHONE_SIMULATOR      - DEPRECATED: Same as TARGET_OS_SIMULATOR
 *    TARGET_OS_NANO               - DEPRECATED: Same as TARGET_OS_WATCH
 *
 *    +--------------------------------------------------------------------------------------+
 *    |                                    TARGET_OS_MAC                                     |
 *    | +-----+ +------------------------------------------------------------+ +-----------+ |
 *    | |     | |                  TARGET_OS_IPHONE                          | |           | |
 *    | |     | | +-----------------+ +----+ +-------+ +--------+ +--------+ | |           | |
 *    | |     | | |       IOS       | |    | |       | |        | |        | | |           | |
 *    | | OSX | | | +-------------+ | | TV | | WATCH | | BRIDGE | | VISION | | | DRIVERKIT | |
 *    | |     | | | | MACCATALYST | | |    | |       | |        | |        | | |           | |
 *    | |     | | | +-------------+ | |    | |       | |        | |        | | |           | |
 *    | |     | | +-----------------+ +----+ +-------+ +--------+ +--------+ | |           | |
 *    | +-----+ +------------------------------------------------------------+ +-----------+ |
 *    +--------------------------------------------------------------------------------------+
 */
```

