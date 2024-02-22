package humanoid_go

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strings"
)

type WordFormatOption int

type LookupMap map[any]any

const (
	LookupIndexPlaceholder                  = iota
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
	wordSetData map[string][]string
	categories  []string
	lookup      map[string]LookupMap
	separator   string
	format      WordFormatOption
	obfuscator  SymmetricObfuscatorInterface
}

type HumanoIDOption func(h *HumanoID) error

func WithCategories(categories []string) HumanoIDOption {
	return func(h *HumanoID) error {
		if len(categories) == 0 {
			return errors.New("categories cannot be empty - remove `WithCategories` call and use autodetect instead")
		}
		h.categories = categories
		return nil
	}
}

func WithSeparator(separator string) HumanoIDOption {
	return func(h *HumanoID) error {
		h.separator = separator
		return nil
	}
}

func WithFormat(format WordFormatOption) HumanoIDOption {
	return func(h *HumanoID) error {
		h.format = format
		return nil
	}
}

func WithObfuscator(obfuscator SymmetricObfuscatorInterface) HumanoIDOption {
	return func(h *HumanoID) error {
		h.obfuscator = obfuscator
		return nil
	}
}

func NewHumanoID(
	wordSets map[string][]string,
	options ...HumanoIDOption,
) (HumanoID, error) {
	humanoid := HumanoID{}
	if len(wordSets) == 0 {
		return humanoid, errors.New("no WordSets provided")
	}

	humanoid.wordSetData = wordSets
	humanoid.categories = make([]string, 0)
	humanoid.lookup = make(map[string]LookupMap)
	humanoid.separator = "-"
	humanoid.obfuscator = NOPObfuscator{}
	// TODO: figure out how to check for categories
	for _, option := range options {
		err := option(&humanoid)
		if err != nil {
			return humanoid, err
		}
	}
	// Assume we need to set up categories
	if options == nil && len(humanoid.categories) == 0 {
		categoryKeys := make([]string, len(humanoid.wordSetData))
		i := 0
		for k, _ := range humanoid.wordSetData {
			categoryKeys[i] = k
			i = i + 1
		}
		humanoid.categories = categoryKeys
	}

	// Build initial lookup table
	// TODO: be aware of a bug if people give custom categories with repeats
	for _, categoryName := range humanoid.categories {
		if len(categoryName) == 0 {
			return humanoid, errors.New(fmt.Sprintf("Category `%s` is invalid", categoryName))
		}
		// TODO: Do the category check and error next

		// Ensure unique and normalized values
		for k, val := range humanoid.wordSetData[categoryName] {
			humanoid.wordSetData[categoryName][k] = strings.ToLower(Trim(val))
		}
		humanoid.wordSetData[categoryName] = uniqueSlice(humanoid.wordSetData[categoryName])

		humanoid.lookup[categoryName] = make(LookupMap)
		for w, i := range sliceToFlippedMap(humanoid.wordSetData[categoryName]) {
			humanoid._addLookup(categoryName, w, i)
		}
	}

	return humanoid, nil
}

func uniqueSlice(slice []string) []string {
	encountered := map[string]bool{}
	result := make([]string, 0)

	for _, v := range slice {
		if encountered[v] == false {
			encountered[v] = true
			result = append(result, v)
		}
	}

	return result
}

func sliceToFlippedMap(mapIn []string) map[string]int {
	newMap := make(map[string]int)
	for k, v := range mapIn {
		newMap[v] = k
	}
	return newMap
}

func (h *HumanoID) Create(id int) (string, error) {
	if id < 0 {
		return "", errors.New("the input ID must be a positive integer")
	}

	// Initialize value to id value
	value := h.obfuscator.Obfuscate(id)
	// Start at last category
	categoryIndex := len(h.categories) - 1
	// Array of words we calculated
	result := make([]string, 0)
	// Get radix
	radix := len(h.wordSetData[h.categories[categoryIndex]])

	for {
		// Determine word for this category
		result = append(result, h._formatWord(h._getWord(categoryIndex, value%radix)))
		// Calculate new value
		value = value / radix
		// Next category (going from highest down to 0, repeating 0 if required)
		categoryIndex = Max(categoryIndex-1, 0)
		// Get radix
		radix = len(h.wordSetData[h.categories[categoryIndex]])

		// break at 0 to replicate do while loop
		if value <= 0 {
			break
		}
	}

	slices.Reverse(result)
	return strings.Join(result, h.separator), nil
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Trim(text string) string {
	return strings.Trim(text, " \n\r\t\v\x00")
}

func (h *HumanoID) Parse(text string) (int, error) {
	value := strings.ToLower(Trim(text))
	if len(value) == 0 {
		return math.MinInt, errors.New("no text specified")
	}

	step := 1
	result := 0
	catIndex := len(h.categories) - 1

	// TODO: PHP does a try/catch here, go no got
	for {
		// Find the index of the word
		wordIndex, _ := h._lookupWordIndex(h.categories[catIndex], value)
		// Add the index * step to the calculated result
		result += wordIndex * step
		// increase step size
		step *= len(h.wordSetData[h.categories[catIndex]])
		// strip found word from text
		// substr($value, 0, -(strlen($this->getWord($catIndex, $ix)) + strlen($this->separator)));
		substrEnd := len(h._getWord(catIndex, wordIndex)) + len(h.separator)
		value = value[:substrEnd]
		catIndex = Max(catIndex-1, 0)

		// LEAVE AT END: replicate "while (value)" in PHP
		if value == "" {
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
	var lastIndex *int
	for _, character := range word {
		if p[string(character)] == nil {
			break
		}

		p := p[string(character)].(LookupMap)
		if index, ok := p[LookupIndexPlaceholder].(*int); ok {
			lastIndex = index
		}
	}

	if lastIndex == nil {
		return math.MinInt, errors.New(fmt.Sprintf("Failed to lookup `%s`", word))
	}

	return *lastIndex, nil
}

func (h *HumanoID) _addLookup(category string, word string, index int) {
	p := h.lookup[category]
	for _, character := range word {
		if p[string(character)] == nil {
			p[string(character)] = make(LookupMap)
			// TODO: probably have to manually save things back to HumanoID?
		}
		p = p[string(character)].(LookupMap)
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

func SpaceIdGenerator(
	options ...HumanoIDOption,
) (HumanoIDInterface, error) {
	var wordSet map[string][]string
	jsonfile, _ := os.Open("./data/space-words.json")
	defer jsonfile.Close()
	dat, _ := io.ReadAll(jsonfile)

	_ = json.Unmarshal([]byte(dat), &wordSet)
	humanoid, err := NewHumanoID(
		wordSet,
		options...,
	)
	return &humanoid, err
}
