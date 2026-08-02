package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type ProbeResult struct {
	DurationSeconds        int
	HasAudio               bool
	Tags                   map[string]string
	AttachedPicStreamIndex int
	HasAttachedPic         bool
}

type M4BOptions struct {
	InputPath              string
	MetadataPath           string
	OutputPath             string
	Bitrate                string
	AttachedPicStreamIndex int
	IncludeCover           bool
	FFmpegPath             string
}

func ProbeInput(inputPath, ffprobePath string) (ProbeResult, error) {
	tags, err := extractTags(inputPath, ffprobePath)
	if err != nil {
		return ProbeResult{}, err
	}
	attachedPicStreamIndex, hasAttachedPic, err := attachedPicStreamIndex(inputPath, ffprobePath)
	if err != nil {
		return ProbeResult{}, err
	}
	hasAudio, err := hasAudioStream(inputPath, ffprobePath)
	if err != nil {
		return ProbeResult{}, err
	}
	durationSeconds, err := durationSeconds(inputPath, ffprobePath)
	if err != nil {
		return ProbeResult{}, err
	}
	return ProbeResult{
		DurationSeconds:        durationSeconds,
		HasAudio:               hasAudio,
		Tags:                   tags,
		AttachedPicStreamIndex: attachedPicStreamIndex,
		HasAttachedPic:         hasAttachedPic,
	}, nil
}

func durationSeconds(inputPath, ffprobePath string) (int, error) {
	cmd := exec.Command(
		ffprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		inputPath,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("Failed to probe input duration: %w; ffprobe stderr: %s", err, stderr.String())
	}

	durationStr := strings.TrimSpace(stdout.String())
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("Failed to parse input duration %q: %w; ffprobe stderr: %s", durationStr, err, stderr.String())
	}

	total := int(duration)
	if duration > float64(total) {
		total++
	}
	return total, nil
}

func hasAudioStream(inputPath, ffprobePath string) (bool, error) {
	cmd := exec.Command(
		ffprobePath,
		"-v", "error",
		"-select_streams", "a",
		"-show_entries", "stream=index",
		"-of", "csv=p=0",
		inputPath,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("Failed to probe input audio stream: %w; ffprobe stderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()) != "", nil
}

func extractTags(inputPath, ffprobePath string) (map[string]string, error) {
	cmd := exec.Command(
		ffprobePath,
		"-v", "error",
		"-print_format", "json",
		"-show_entries", "format_tags",
		inputPath,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Failed to extract input tags: %w; ffprobe stderr: %s", err, stderr.String())
	}

	var payload struct {
		Format struct {
			Tags map[string]string `json:"tags"`
		} `json:"format"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		return nil, fmt.Errorf("Failed to parse ffprobe metadata output: %w; ffprobe stdout: %s; stderr: %s", err, stdout.String(), stderr.String())
	}

	// map common source tag keys to ffmetadata/global keys used by m4b
	tagKeyMap := map[string]string{
		"title":        "title",
		"artist":       "artist",
		"author":       "artist",
		"album_artist": "album_artist",
		"album":        "album",
		"genre":        "genre",
		"date":         "date",
		"year":         "date",
		"track":        "track",
		"disc":         "disc",
		"composer":     "composer",
		"comment":      "comment",
		"description":  "description",
		"publisher":    "publisher",
		"copyright":    "copyright",
	}

	result := make(map[string]string)
	for k, v := range payload.Format.Tags {
		srcKey := strings.ToLower(strings.TrimSpace(k))
		dstKey, ok := tagKeyMap[srcKey]
		if !ok {
			continue
		}
		value := strings.TrimSpace(v)
		if value == "" {
			continue
		}
		if _, exists := result[dstKey]; exists {
			continue
		}
		result[dstKey] = value
	}

	return result, nil
}

func attachedPicStreamIndex(inputPath, ffprobePath string) (int, bool, error) {
	cmd := exec.Command(
		ffprobePath,
		"-v", "error",
		"-print_format", "json",
		"-select_streams", "v",
		"-show_entries", "stream=index:stream_disposition=attached_pic",
		inputPath,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return -1, false, fmt.Errorf("Failed to inspect cover art stream: %w; ffprobe stderr: %s", err, stderr.String())
	}

	var payload struct {
		Streams []struct {
			Index       int `json:"index"`
			Disposition struct {
				AttachedPic int `json:"attached_pic"`
			} `json:"disposition"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		return -1, false, fmt.Errorf("Failed to parse cover art stream metadata: %w; ffprobe stdout: %s; stderr: %s", err, stdout.String(), stderr.String())
	}

	for _, s := range payload.Streams {
		if s.Disposition.AttachedPic == 1 {
			return s.Index, true, nil
		}
	}

	return -1, false, nil
}

func CreateM4B(opts M4BOptions) error {
	bitrate := opts.Bitrate
	if bitrate == "" {
		bitrate = "96k"
	}
	outputExt := filepath.Ext(opts.OutputPath)

	args := []string{
		"-y",
		"-i", opts.InputPath,
		"-f", "ffmetadata",
		"-i", opts.MetadataPath,
		"-map", "0:a",
		"-map_metadata", "1",
		"-map_chapters", "1",
		"-c:a", "aac",
		"-b:a", bitrate,
	}

	if opts.IncludeCover && opts.AttachedPicStreamIndex >= 0 {
		args = append(args,
			"-map", fmt.Sprintf("0:%d", opts.AttachedPicStreamIndex),
			"-c:v", "copy",
			"-disposition:v:0", "attached_pic",
		)
		if strings.EqualFold(outputExt, ".mp3") {
			args = append(args, "-id3v2_version", "3")
		}
	}
	args = append(args, opts.OutputPath)

	cmd := exec.Command(opts.FFmpegPath, args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to create m4b output: %w; ffmpeg output: %s",
			err,
			out.String(),
		)
	}
	return nil
}
