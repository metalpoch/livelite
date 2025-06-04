package ffmpeg

import (
	"os"
	"os/exec"
	"path/filepath"
)

func StreamToHLS(rtmpURL, hlsSegmentFilename, hlsPath, hlsTime, hlsListSize string) error {
	dir := filepath.Dir(hlsSegmentFilename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return exec.Command("ffmpeg",
		"-y",
		"-i", rtmpURL,
		"-c:v", "copy", "-c:a", "copy",
		// "-c:v", "libx264", "-preset", "veryfast", "-c:a", "aac", // video a H.264 y el audio a AAC, ambos compatibles con HLS
		"-f", "hls",
		"-hls_time", hlsTime,
		"-hls_list_size", hlsListSize,
		"-hls_playlist_type", "event",
		"-hls_segment_filename", hlsSegmentFilename,
		hlsPath,
	).Start()
}
