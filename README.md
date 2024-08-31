# ASCII Video Player
**ASCII Video Player** is a terminal-based video player that transforms videos into ASCII art.

![ASCII Video Player Demo 1](docs/images/lion.png)
![ASCII Video Player Demo 2](docs/images/goku.png)

## Table of Contents

- [How to Use](#how-to-use)
- [Functionality](#functionality)
- [Example GIFs](#example-gifs)
- [Supported Platforms](#supported-platforms)
- [Dependencies](#dependencies)
- [Build From Source](#build)
- [Development Setup](#development-setup)

## How to Use <a name="how-to-use"></a>

To play a video in the terminal, simply run the following command:

```bash
ascii-player --video path/to/video.mp4
```

## Functionality <a name="functionality"></a>

- [x] Supports RGB colors / full greyscale
- [x] Supports audio playback
- [x] Compatible with any resolution/framerate, automatically downscales to fit the terminal
- [] Simple pause/resume video controls using the spacebar (TODO)
- [] Directly supports YouTube URLs (TODO)

## Example GIFs <a name="example-gifs"></a>

## Supported Platforms <a name="supported-platforms"></a>
- Linux

## Dependencies <a name="dependencies"></a>

- **FFmpeg**: For video decoding and processing.
- **MediaInfo**: For extracting video metadata.

## Build from source <a name="build-from-source"></a>
```bash
make build
```

## Development Setup <a name="development-setup"></a>

To install the development dependencies, run:

```bash
make install-dev
```

This command installs FFmpeg, MediaInfo, and all Go module dependencies required for development.
