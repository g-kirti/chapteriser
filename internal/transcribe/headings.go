package transcribe

import (
	"encoding/json"
	"g-kirti/chapteriser/internal/domain"
	"log"
	"strings"
)

const maxChapterNumberWordGapSeconds = 0.45
const minStandaloneHeadingFollowingGapSeconds = 0.6
const standaloneHeadingNearBoundarySeconds = 45.0

const (
	numUnknown = iota
	numUnit
	numTens
	numScale
	numAnd
)

func numberWordType(word string) int {
	switch word {
	case "zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine",
		"ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen",
		"eighteen", "nineteen":
		return numUnit
	case "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety":
		return numTens
	case "hundred", "thousand":
		return numScale
	case "and":
		return numAnd
	default:
		return numUnknown
	}
}

func canAppendNumberWord(prevWord string, nextWord string) bool {
	prevType := numberWordType(prevWord)
	nextType := numberWordType(nextWord)
	if nextType == numUnknown {
		return false
	}
	if nextType == numAnd {
		return prevType == numScale
	}
	if prevType == numAnd {
		return nextType == numUnit || nextType == numTens
	}
	if prevType == numTens {
		return nextType == numUnit
	}
	if prevType == numUnit {
		return nextType == numScale
	}
	if prevType == numScale {
		return nextType == numUnit || nextType == numTens || nextType == numAnd
	}
	return false
}

func canStartChapterNumber(word string) bool {
	t := numberWordType(word)
	return t == numUnit || t == numTens
}

func shouldAcceptStandaloneHeading(words []struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}, idx int) bool {
	if idx < 0 || idx >= len(words) {
		return false
	}
	curr := words[idx]

	if curr.Start <= standaloneHeadingNearBoundarySeconds {
		return true
	}
	if idx == len(words)-1 {
		return true
	}
	next := words[idx+1]
	if next.Start-curr.End >= minStandaloneHeadingFollowingGapSeconds {
		return true
	}
	return false
}

func FindHeadings(resultJSON string) []domain.Chapter {
	var result struct {
		Result []struct {
			Word  string  `json:"word"`
			Start float64 `json:"start"`
			End   float64 `json:"end"`
		} `json:"result"`
	}

	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		log.Println("[transcribe] Failed to parse JSON:", err)
		return nil
	}

	keywords := map[string]bool{
		"book":    true,
		"chapter": true,
		"volume":  true,
		"part":    true,
	}
	standaloneKeywords := map[string]bool{
		"introduction": true,
		"prologue":     true,
		"epilogue":     true,
		"foreword":     true,
		"afterword":    true,
	}

	numberWords := map[string]bool{
		"zero": true, "one": true, "two": true, "three": true, "four": true,
		"five": true, "six": true, "seven": true, "eight": true, "nine": true,
		"ten": true, "eleven": true, "twelve": true, "thirteen": true, "fourteen": true,
		"fifteen": true, "sixteen": true, "seventeen": true, "eighteen": true, "nineteen": true,
		"twenty": true, "thirty": true, "forty": true, "fifty": true, "sixty": true,
		"seventy": true, "eighty": true, "ninety": true, "hundred": true,
		"thousand": true, "and": true,
	}

	var chapterList []domain.Chapter

	for i := 0; i < len(result.Result); i++ {
		curr := result.Result[i]
		if standaloneKeywords[curr.Word] && shouldAcceptStandaloneHeading(result.Result, i) {
			log.Printf("[transcribe] Found '%s' at [%.2fs - %.2fs]\n", curr.Word, curr.Start, curr.End)
			chapterList = append(chapterList, domain.Chapter{
				Title: curr.Word,
				Start: curr.Start,
				End:   curr.End,
			})
			continue
		}
		if !keywords[curr.Word] {
			continue
		}

		phrase := []string{curr.Word}
		startTime := curr.Start
		endTime := curr.End

		j := i + 1
		for j < len(result.Result) {
			prev := result.Result[j-1]
			next := result.Result[j]
			if next.Start-prev.End > maxChapterNumberWordGapSeconds {
				break
			}
			if len(phrase) == 1 {
				if numberWords[next.Word] && canStartChapterNumber(next.Word) {
					phrase = append(phrase, next.Word)
					endTime = next.End
					j++
					continue
				}
				break
			}
			lastNumberWord := phrase[len(phrase)-1]
			if numberWords[next.Word] && canAppendNumberWord(lastNumberWord, next.Word) {
				phrase = append(phrase, next.Word)
				endTime = next.End
				j++
			} else {
				break
			}
		}

		if len(phrase) > 1 {
			joinedPhrase := strings.Join(phrase, " ")
			log.Printf("[transcribe] Found '%s' at [%.2fs - %.2fs]\n", strings.Join(phrase, " "), startTime, endTime)

			chapterList = append(chapterList, domain.Chapter{
				Title: joinedPhrase,
				Start: startTime,
				End:   endTime,
			})
		}
		i = j - 1
	}
	return chapterList
}
