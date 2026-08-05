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
	numTeenBase
)

func numberWordType(word string, wordTypes map[string]int) int {
	if t, ok := wordTypes[word]; ok {
		return t
	}
	return numUnknown
}

func canAppendNumberWord(prevWord string, nextWord string, wordTypes map[string]int, andFollowsTypes map[int]bool) bool {
	prevType := numberWordType(prevWord, wordTypes)
	nextType := numberWordType(nextWord, wordTypes)
	if nextType == numUnknown {
		return false
	}
	if nextType == numAnd {
		return andFollowsTypes[prevType]
	}
	if prevType == numAnd {
		return nextType == numUnit || nextType == numTens
	}
	if prevType == numUnit {
		return nextType == numScale
	}
	if prevType == numTens {
		return nextType == numUnit || nextType == numScale || nextType == numTeenBase
	}
	if prevType == numScale {
		return nextType == numUnit || nextType == numTens || nextType == numScale
	}
	if prevType == numTeenBase {
		return nextType == numUnit || nextType == numScale
	}
	return false
}

func canStartChapterNumber(word string, wordTypes map[string]int) bool {
	t := numberWordType(word, wordTypes)
	return t == numUnit || t == numTens || t == numScale || t == numTeenBase
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

func FindHeadings(resultJSON string, langCode string) []domain.Chapter {
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

	lang := LanguageRegistry[langCode]
	keywords := lang.Keywords
	standaloneKeywords := lang.StandaloneKeywords
	numberWordTypes := lang.NumberWordTypes
	andFollowsTypes := lang.AndFollowsTypes

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
				if canStartChapterNumber(next.Word, numberWordTypes) {
					phrase = append(phrase, next.Word)
					endTime = next.End
					j++
					continue
				}
				break
			}
			lastNumberWord := phrase[len(phrase)-1]
			if canAppendNumberWord(lastNumberWord, next.Word, numberWordTypes, andFollowsTypes) {
				phrase = append(phrase, next.Word)
				endTime = next.End
				j++
			} else {
				break
			}
		}
		if len(phrase) > 1 {
			joinedPhrase := strings.Join(phrase, " ")
			log.Printf("[transcribe] Found '%s' at [%.2fs - %.2fs]\n", joinedPhrase, startTime, endTime)
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
