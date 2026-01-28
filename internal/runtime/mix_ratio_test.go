package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMixRatio_Valid(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name     string
		input    string
		expected []mixRatio
	}{
		{
			name:  "three-way split",
			input: "REALM_CALL:70,REALM_DEPLOYMENT:20,PACKAGE_DEPLOYMENT:10",
			expected: []mixRatio{
				{RealmCall, 70},
				{RealmDeployment, 20},
				{PackageDeployment, 10},
			},
		},
		{
			name:  "two-way split",
			input: "REALM_CALL:50,REALM_DEPLOYMENT:50",
			expected: []mixRatio{
				{RealmCall, 50},
				{RealmDeployment, 50},
			},
		},
		{
			name:  "with spaces",
			input: "REALM_CALL:70, REALM_DEPLOYMENT:30",
			expected: []mixRatio{
				{RealmCall, 70},
				{RealmDeployment, 30},
			},
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, err := ParseMixRatio(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, config.Ratios)
		})
	}
}

func TestParseMixRatio_Invalid(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		expectedErr error
		name        string
		input       string
	}{
		{
			name:        "empty string",
			input:       "",
			expectedErr: errEmptyMixRatio,
		},
		{
			name:        "whitespace only",
			input:       "   ",
			expectedErr: errEmptyMixRatio,
		},
		{
			name:        "single type",
			input:       "REALM_CALL:100",
			expectedErr: errInsufficientTypes,
		},
		{
			name:        "missing colon",
			input:       "REALM_CALL70,REALM_DEPLOYMENT30",
			expectedErr: errInvalidRatioFormat,
		},
		{
			name:        "invalid percentage - negative",
			input:       "REALM_CALL:-10,REALM_DEPLOYMENT:110",
			expectedErr: errInvalidPercentage,
		},
		{
			name:        "invalid percentage - zero",
			input:       "REALM_CALL:0,REALM_DEPLOYMENT:100",
			expectedErr: errInvalidPercentage,
		},
		{
			name:        "invalid percentage - over 100",
			input:       "REALM_CALL:101,REALM_DEPLOYMENT:0",
			expectedErr: errInvalidPercentage,
		},
		{
			name:        "invalid percentage - not a number",
			input:       "REALM_CALL:abc,REALM_DEPLOYMENT:50",
			expectedErr: errInvalidPercentage,
		},
		{
			name:        "unknown type",
			input:       "UNKNOWN_TYPE:50,REALM_DEPLOYMENT:50",
			expectedErr: errUnknownType,
		},
		{
			name:        "duplicate type",
			input:       "REALM_CALL:50,REALM_CALL:50",
			expectedErr: errDuplicateType,
		},
		{
			name:        "MIXED in mix",
			input:       "MIXED:50,REALM_CALL:50",
			expectedErr: errMixedInMix,
		},
		{
			name:        "sum not 100 - under",
			input:       "REALM_CALL:40,REALM_DEPLOYMENT:40",
			expectedErr: errRatioSumNot100,
		},
		{
			name:        "sum not 100 - over",
			input:       "REALM_CALL:60,REALM_DEPLOYMENT:60",
			expectedErr: errRatioSumNot100,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseMixRatio(tc.input)
			require.Error(t, err)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestMixConfig_HasType(t *testing.T) {
	t.Parallel()

	config := &MixConfig{
		Ratios: []mixRatio{
			{RealmCall, 70},
			{RealmDeployment, 30},
		},
	}

	assert.True(t, config.HasType(RealmCall))
	assert.True(t, config.HasType(RealmDeployment))
	assert.False(t, config.HasType(PackageDeployment))
	assert.False(t, config.HasType(Mixed))
}

func TestMixConfig_CalculateTransactionCounts(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name     string
		expected map[Type]uint64
		ratios   []mixRatio
		total    uint64
	}{
		{
			name: "exact division",
			ratios: []mixRatio{
				{RealmCall, 70},
				{RealmDeployment, 20},
				{PackageDeployment, 10},
			},
			total: 100,
			expected: map[Type]uint64{
				RealmCall:         70,
				RealmDeployment:   20,
				PackageDeployment: 10,
			},
		},
		{
			name: "with rounding - remainder goes to last",
			ratios: []mixRatio{
				{RealmCall, 70},
				{RealmDeployment, 30},
			},
			total: 10,
			expected: map[Type]uint64{
				RealmCall:       7,
				RealmDeployment: 3,
			},
		},
		{
			name: "small total with rounding",
			ratios: []mixRatio{
				{RealmCall, 33},
				{RealmDeployment, 33},
				{PackageDeployment, 34},
			},
			total: 10,
			expected: map[Type]uint64{
				RealmCall:         3,
				RealmDeployment:   3,
				PackageDeployment: 4,
			},
		},
		{
			name: "two-way 50/50",
			ratios: []mixRatio{
				{RealmCall, 50},
				{PackageDeployment, 50},
			},
			total: 100,
			expected: map[Type]uint64{
				RealmCall:         50,
				PackageDeployment: 50,
			},
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := &MixConfig{Ratios: tc.ratios}
			counts := config.CalculateTransactionCounts(tc.total)
			assert.Equal(t, tc.expected, counts)
		})
	}
}
