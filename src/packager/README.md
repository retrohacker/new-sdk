# Packager

Generate a self-expanding executable from the StorjSDK binary

# Technical Overview

This is a small go utility that takes the StorjSDK binary, places a byte pattern "marker" at the end, and then packs a tar.gz file into the end. This allows us to distribute a single executable as the "StorjSDK binary" for every platform, and all dependent files can be unpacked at runtime.

Note that this is not magic. The StorjSDK binary needs to be written to detect the magic "marker" at the end of the executable and unpack the tar.gz itself.
