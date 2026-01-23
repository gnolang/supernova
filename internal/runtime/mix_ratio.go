package runtime

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	errEmptyMixRatio      = errors.New("mix ratio cannot be empty")
	errInvalidRatioFormat = errors.New("invalid ratio format, expected TYPE:PERCENTAGE")
	errInvalidPercentage  = errors.New("percentage must be a positive integer between 1 and 100")
	errUnknownType        = errors.New("unknown runtime type in mix ratio")
	errDuplicateType      = errors.New("duplicate runtime type in mix ratio")
	errMixedInMix         = errors.New("MIXED type cannot be used in mix ratio")
	errRatioSumNot100     = errors.New("mix ratio percentages must sum to 100")
	errInsufficientTypes  = errors.New("mix ratio must contain at least 2 types")
)

type MixRatio struct {
	Type       Type
	Percentage int
}

type MixConfig struct {
	Ratios []MixRatio
}

// ParseMixRatio parses a mix ratio string into a MixConfig
// Example input: "REALM_CALL:70,REALM_DEPLOYMENT:20,PACKAGE_DEPLOYMENT:10"
func ParseMixRatio(input string) (*MixConfig, error) {
	if strings.TrimSpace(input) == "" {
		return nil, errEmptyMixRatio
	}

	parts := strings.Split(input, ",")
	if len(parts) < 2 {
		return nil, errInsufficientTypes
	}

	config := &MixConfig{
		Ratios: make([]MixRatio, 0, len(parts)),
	}

	seenTypes := make(map[Type]bool)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		ratio, err := parseRatioPart(part)
		if err != nil {
			return nil, err
		}

		if seenTypes[ratio.Type] {
			return nil, fmt.Errorf("%w: %s", errDuplicateType, ratio.Type)
		}
		seenTypes[ratio.Type] = true

		config.Ratios = append(config.Ratios, ratio)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// parseRatioPart parses a single TYPE:PERCENTAGE part of the mix ratio
// Example input: "REALM_CALL:70" and ensures validity
func parseRatioPart(part string) (MixRatio, error) {
	colonIdx := strings.LastIndex(part, ":")
	if colonIdx == -1 {
		return MixRatio{}, fmt.Errorf("%w: %s", errInvalidRatioFormat, part)
	}

	typeName := strings.TrimSpace(part[:colonIdx])
	percentageStr := strings.TrimSpace(part[colonIdx+1:])

	percentage, err := strconv.Atoi(percentageStr)
	if err != nil || percentage < 1 || percentage > 100 {
		return MixRatio{}, fmt.Errorf("%w: %s", errInvalidPercentage, percentageStr)
	}

	runtimeType := Type(typeName)

	if runtimeType == Mixed {
		return MixRatio{}, errMixedInMix
	}

	if !IsMixableRuntime(runtimeType) {
		return MixRatio{}, fmt.Errorf("%w: %s", errUnknownType, typeName)
	}

	return MixRatio{
		Type:       runtimeType,
		Percentage: percentage,
	}, nil
}

// Validate checks the MixConfig for correctness
// ensuring at least two types and that percentages sum to 100
func (mc *MixConfig) Validate() error {
	if len(mc.Ratios) < 2 {
		return errInsufficientTypes
	}

	sum := 0
	for _, ratio := range mc.Ratios {
		sum += ratio.Percentage
	}

	if sum != 100 {
		return fmt.Errorf("%w: got %d", errRatioSumNot100, sum)
	}

	return nil
}

func (mc *MixConfig) HasType(t Type) bool {
	for _, ratio := range mc.Ratios {
		if ratio.Type == t {
			return true
		}
	}
	return false
}

// CalculateTransactionCounts computes the number of transactions
// for each runtime type based on the total and the defined ratios
func (mc *MixConfig) CalculateTransactionCounts(total uint64) map[Type]uint64 {
	counts := make(map[Type]uint64)
	var allocated uint64

	for i, ratio := range mc.Ratios {
		if i == len(mc.Ratios)-1 {
			counts[ratio.Type] = total - allocated
		} else {
			count := (total * uint64(ratio.Percentage)) / 100
			counts[ratio.Type] = count
			allocated += count
		}
	}

	return counts
}
