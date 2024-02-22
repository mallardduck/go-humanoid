package humanoid_go

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

type WordFormatOption int

type LookupMap map[any]any

const (
	LookupIndexPlaceholder                  = 0
	UpperCaseFirst         WordFormatOption = iota
	LowerCaseFirst
	UpperCase
	LowerCase
)

type HumanoIDInterface interface {
	Create(id int) (string, error)
	Parse(text string) (int, error)
}

type SymmetricObfuscatorInterface interface {
	Obfuscate(id int) int
	Deobfuscate(id int) int
}

type HumanoID struct {
	wordSetData map[string]map[int]string
	categories  map[string]int
	lookup      map[string]LookupMap
	separator   string
	format      WordFormatOption
	obfuscator  SymmetricObfuscatorInterface
}

type HumanoIDOption func(h *HumanoID)

func WithCategories(categories map[int]string) HumanoIDOption {
	return func(h *HumanoID) {
		h.categories = categories
	}
}

func WithSeparator(separator string) HumanoIDOption {
	return func(h *HumanoID) {
		h.separator = separator
	}
}

func WithFormat(format WordFormatOption) HumanoIDOption {
	return func(h *HumanoID) {
		h.format = format
	}
}

func WithObfuscator(obfuscator SymmetricObfuscatorInterface) HumanoIDOption {
	return func(h *HumanoID) {
		h.obfuscator = obfuscator
	}
}

func NewHumanoID(
	wordSets map[string]map[int]string,
	opts ...HumanoIDOption,
) (HumanoID, error) {
	if len(wordSets) == 0 {
		return _, errors.New("no WordSets provided")
	}

	// TODO: figure out how to make opts do the things under this..

	humanoid := HumanoID{}
	humanoid.wordSetData = wordSets

	catKeys := make(map[int]string, len(wordSets))
	for k, _ := range wordSets {
		catKeys = append(catKeys, k)
	}
	humanoid.categories = catKeys

	for _, categoryName := range categories {
		// TODO: construct initial look up map
	}
	humanoid.separator = separator
	humanoid.format = format
	humanoid.obfuscator = NOPObfuscator{}
	return humanoid
}

func (h *HumanoID) Create(id int) (string, error) {
	if id < 0 {
		return "", errors.New("the input ID must be a positive integer")
	}

	value := h.obfuscator.Obfuscate(id)
	categoryIndex := len(h.categories) - 1
	result := make(map[int]string)
	radix := len(h.wordSetData[h.categories[categoryIndex]])

	for {
		// Determine word for this category
		result = append(result, h._formatWord(categoryIndex, value % radix))
		// Calculate new value
		// Todo: make this an in
		value = value / radix
		// Next category (going from highest down to 0, repeating 0 if required)
		categoryIndex = math.Max(--categoryIndex, 0)
		// Get radix
		radix = len(h.wordSetData[h.categories[categoryIndex]])

		// break at 0 to replicate do while loop
		if value <= 0 {
			break
		}
	}

	// TODO: reverse results
	return strings.Join(result, h.separator), _
}

func (h *HumanoID) Parse(text string) (int, error) {
	value := strings.ToLower(strings.Trim(text, " \n\r\t\v\x00"))
	if len(value) == 0 {
		return math.MinInt, errors.New("no text specified")
	}

	step := 1
	result := 0
	catIndex := len(h.categories) - 1

	// TODO: PHP does a try/catch here, go no got
	for {
		// Find the index of the word
		wordIndex, err := h._lookupWordIndex(h.categories[catIndex], value)
		// Add the index * step to the calculated result
		result += (wordIndex * step)
		// increase step size
		step *= h.wordSetData[h.categories][catIndex]
		// strip found word from text
		// substr($value, 0, -(strlen($this->getWord($catIndex, $ix)) + strlen($this->separator)));
		substrEnd := len(h._getWord(catIndex, wordIndex)) + len(h.separator)
		value = value[:substrEnd]
		catIndex = math.Max(--catIndex, 0)

		// LEAVE AT END: replicate "while (value)" in PHP
		if value == nil {
			break
		}
	}
	// TODO: some where between here and todo above, error handling

	return h.obfuscator.Deobfuscate(result), nil
}

func (h *HumanoID) _getWord(categoryIndex int, wordIndex int) string {
	return h.wordSetData[h.categories[categoryIndex]][wordIndex]
}

func (h *HumanoID) _formatWord(word string) string {
	switch h.format {
	case UpperCaseFirst:
		return strings.ToUpper(word[0:1]) + word[1:]
	case LowerCaseFirst:
		return strings.ToLower(word[0:1]) + word[1:]
	case UpperCase:
		return strings.ToUpper(word)
	case LowerCase:
		return strings.ToLower(word)
	default:
		return word
	}
}

func (h *HumanoID) _lookupWordIndex(category string, word string) (int, error) {
	p := h.lookup[category]
	var lastIndex int
	for i, character := range word {
		_, err := p[character]
		if err {
			break
		}

		p := &p[character]
		_, err := p[LookupIndexPlaceholder]
		if !err {
			lastIndex = p[LookupIndexPlaceholder]
		}
	}

	if lastIndex == nil {
		return math.MinInt, errors.New(fmt.Sprintf("Failed to lookup `%s`", word))
	}

	return lastIndex, nil
}

func (h *HumanoID) _addLookup(category string, word string, index int) {
	p := h.lookup[category]
	for i, c := range word {
		_, err := p[c]
		if err {
			p[c] = make(LookupMap)
		}
		p = &p[c]
	}

	p[LookupIndexPlaceholder] = index
}

type NOPObfuscator struct{}

func (N NOPObfuscator) Obfuscate(id int) int {
	return id
}

func (N NOPObfuscator) Deobfuscate(id int) int {
	return id
}

type BasicShiftObfuscator struct {
	salt int
}

func (O BasicShiftObfuscator) Obfuscate(id int) int {
	return id ^ O.salt
}

func (O BasicShiftObfuscator) Deobfuscate(id int) int {
	return id ^ O.salt
}
