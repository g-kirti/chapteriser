package metadata

import (
	"fmt"
	"g-kirti/chapteriser/internal/domain"
	"os"
	"sort"
	"strings"
	"unicode"
)

func WriteChaptersToFile(chapters []domain.Chapter, filename string, tags map[string]string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(";FFMETADATA1\n")
	writeGlobalTags(f, tags)

	for i := range len(chapters) {
		start := int(chapters[i].Start * 1000)
		var end int
		if i < len(chapters)-1 {
			end = int(chapters[i+1].Start * 1000)
		} else {
			end = int((chapters[i].Start + 300) * 1000)
		}

		fmt.Fprintf(f, "[CHAPTER]\n")
		fmt.Fprintf(f, "TIMEBASE=1/1000\n")
		fmt.Fprintf(f, "START=%d\n", start)
		fmt.Fprintf(f, "END=%d\n", end)
		fmt.Fprintf(f, "title=%s\n\n", escapeFFMetadata(titleCase(chapters[i].Title)))
	}
	return nil
}

func writeGlobalTags(f *os.File, tags map[string]string) {
	if len(tags) == 0 {
		return
	}

	keys := make([]string, 0, len(tags))
	for k, v := range tags {
		if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(f, "%s=%s\n", k, escapeFFMetadata(tags[k]))
	}
	fmt.Fprintln(f)
}

func escapeFFMetadata(s string) string {
	replacer := strings.NewReplacer(
		`\\`, `\\\\`,
		";", `\;`,
		"#", `\#`,
		"=", `\=`,
		"\n", " ",
		"\r", " ",
	)
	return replacer.Replace(s)
}

func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		runes := []rune(w)
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		for j := 1; j < len(runes); j++ {
			runes[j] = unicode.ToLower(runes[j])
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}
