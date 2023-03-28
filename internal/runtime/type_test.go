package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType_IsRuntime(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name    string
		mode    Type
		isValid bool
	}{
		{
			"Realm Deployment",
			RealmDeployment,
			true,
		},
		{
			"Package Deployment",
			PackageDeployment,
			true,
		},
		{
			"Realm Call",
			RealmCall,
			true,
		},
		{
			"Dummy mode",
			Type("Dummy mode"),
			false,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.isValid, IsRuntime(testCase.mode))
		})
	}
}

func TestType_String(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name        string
		mode        Type
		expectedStr string
	}{
		{
			"Realm Deployment",
			RealmDeployment,
			string(RealmDeployment),
		},
		{
			"Package Deployment",
			PackageDeployment,
			string(PackageDeployment),
		},
		{
			"Realm Call",
			RealmCall,
			string(RealmCall),
		},
		{
			"Dummy mode",
			Type("Dummy mode"),
			string(unknown),
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.expectedStr, testCase.mode.String())
		})
	}
}
