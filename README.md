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
- [Development Setup](#development-setup)

## How to Use

To play a video in the terminal, simply run the following command:

```bash
go run main.go --video path/to/video.mp4
```

Replace `path/to/video.mp4` with the path to your desired video file or URL.

## Functionality

- [x] Supports RGB colors
- [x] Supports audio playback
- [x] Simple pause/resume video controls using the spacebar
- [x] Directly supports YouTube URLs
- [x] Compatible with any resolution/framerate, automatically downscales to fit the terminal

## Example GIFs

## Supported Platforms
- Linux

## Dependencies

- **FFmpeg**: For video decoding and processing.
- **MediaInfo**: For extracting video metadata.

## Development Setup

To install the development dependencies, run:

```bash
make install-dev
```

This command installs FFmpeg, MediaInfo, and all Go module dependencies required for development.

