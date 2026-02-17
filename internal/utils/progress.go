package utils

import (
	"bufio"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type ProgressParser struct {
	regex              *regexp.Regexp
	currentFormat      string
	currentDestination string
}

func NewProgressParser() *ProgressParser {
	return &ProgressParser{
		regex: regexp.MustCompile(`(\d+(?:\.\d+)?)%`),
	}
}

func (p *ProgressParser) ReadPipe(pipe io.Reader, sendProgress func(float64, string, string, string, string)) {
	reader := bufio.NewReader(pipe)
	var lineBuilder strings.Builder

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			if lineBuilder.Len() > 0 {
				line := lineBuilder.String()
				percent, speed, eta, status, destination := p.ParseLine(line)
				if strings.Contains(line, "[download]") || percent > 0 || speed != "" || eta != "" {
					sendProgress(percent, speed, eta, status, destination)
				}
			}

			break
		}

		switch r {
		// TODO: test this on windows and remove if not needed
		case '\r':
			if lineBuilder.Len() > 0 {
				line := lineBuilder.String()
				percent, speed, eta, status, destination := p.ParseLine(line)
				if strings.Contains(line, "[download]") || percent > 0 || speed != "" || eta != "" {
					log.Printf("Progress parsed (\\r): %.2f%%, speed: %s, eta: %s, status: %s, destination: %s, line: %s", percent, speed, eta, status, destination, line)
					sendProgress(percent, speed, eta, status, destination)
				}

				lineBuilder.Reset()
			}

		case '\n':
			if lineBuilder.Len() > 0 {
				line := lineBuilder.String()
				percent, speed, eta, status, destination := p.ParseLine(line)
				if strings.Contains(line, "[download]") || percent > 0 || speed != "" || eta != "" {
					sendProgress(percent, speed, eta, status, destination)
				}
				lineBuilder.Reset()
			}

		default:
			lineBuilder.WriteRune(r)
		}
	}
}

func (p *ProgressParser) ParseLine(line string) (percent float64, speed, eta, status, destination string) {
	p.currentDestination = ""
	p.currentFormat = ""

	percentPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(\d+(?:\.\d+)?)\s*%`),
		regexp.MustCompile(`\[download\]\s+(\d+(?:\.\d+)?)%`),
	}

	for _, pattern := range percentPatterns {
		percentMatch := pattern.FindStringSubmatch(line)
		if len(percentMatch) > 1 {
			if pr, err := strconv.ParseFloat(percentMatch[1], 64); err == nil {
				percent = pr
				break
			}
		}
	}

	speedPattern := regexp.MustCompile(`(\d+(?:\.\d+)?[KMG]?i?B/s)`)
	speedMatch := speedPattern.FindStringSubmatch(line)
	if len(speedMatch) > 1 {
		speed = speedMatch[1]
	}

	etaPattern := regexp.MustCompile(`ETA\s+(\d+:\d+(?::\d+)?)`)
	etaMatch := etaPattern.FindStringSubmatch(line)
	if len(etaMatch) > 1 {
		eta = etaMatch[1]
	}

	if strings.Contains(line, "[download] Destination:") {
		destPattern := regexp.MustCompile(`Destination:\s*(.+)`)
		if match := destPattern.FindStringSubmatch(line); len(match) > 1 {
			p.currentDestination = strings.TrimSpace(match[1])
		}

		if ext := extractFormatFromDestination(line); ext != "" {
			p.currentFormat = ext
		}
	}

	formatPattern := regexp.MustCompile(`(?:format|format_id)\s+(\d+)`)
	if match := formatPattern.FindStringSubmatch(line); len(match) > 1 {
		p.currentFormat = "format " + match[1]
	}

	if percent > 0 {
		if p.currentFormat != "" {
			status = "[download] " + p.currentFormat
		} else {
			status = "[download]"
		}
	}

	return percent, speed, eta, status, p.currentDestination
}

func extractFormatFromDestination(line string) string {
	videoExtensions := map[string]bool{
		".mp4":  true,
		".webm": true,
		".mkv":  true,
	}
	audioExtensions := map[string]bool{
		".m4a":  true,
		".mp3":  true,
		".ogg":  true,
		".wav":  true,
		".flac": true,
		".aac":  true,
	}

	for ext := range videoExtensions {
		if strings.Contains(line, ext) {
			return "video"
		}
	}

	for ext := range audioExtensions {
		if strings.Contains(line, ext) {
			return "audio"
		}
	}

	return ""
}
