package cli

import (
	"testing"

	"github.com/ihsanmokhlisse/gitopsi/internal/operator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperatorCmd_Exists(t *testing.T) {
	assert.NotNil(t, operatorCmd, "operatorCmd should exist")
	assert.Equal(t, "operator", operatorCmd.Use)
}

func TestOperatorAddCmd_Exists(t *testing.T) {
	assert.NotNil(t, operatorAddCmd, "operatorAddCmd should exist")
	assert.Equal(t, "add [name]", operatorAddCmd.Use)
}

func TestOperatorAddPresetCmd_Exists(t *testing.T) {
	assert.NotNil(t, operatorAddPresetCmd, "operatorAddPresetCmd should exist")
	assert.Equal(t, "add-preset [preset]", operatorAddPresetCmd.Use)
}

func TestOperatorRemoveCmd_Exists(t *testing.T) {
	assert.NotNil(t, operatorRemoveCmd, "operatorRemoveCmd should exist")
	assert.Equal(t, "remove [name]", operatorRemoveCmd.Use)
}

func TestOperatorListCmd_Exists(t *testing.T) {
	assert.NotNil(t, operatorListCmd, "operatorListCmd should exist")
	assert.Equal(t, "list", operatorListCmd.Use)
}

func TestOperatorPresetsCmd_Exists(t *testing.T) {
	assert.NotNil(t, operatorPresetsCmd, "operatorPresetsCmd should exist")
	assert.Equal(t, "presets", operatorPresetsCmd.Use)
}

func TestGetOperatorManager_ReturnsManager(t *testing.T) {
	tmpDir := t.TempDir()
	operatorProjectPath = tmpDir

	mgr, err := getOperatorManager()
	require.NoError(t, err, "getOperatorManager should not error")
	assert.NotNil(t, mgr, "Manager should not be nil")
}

func TestOperatorPresets_ListsAllPresets(t *testing.T) {
	presets := operator.ListOperatorPresets()
	assert.True(t, len(presets) > 0, "Should have presets available")
	assert.Contains(t, presets, "prometheus", "Should include prometheus preset")
	assert.Contains(t, presets, "grafana", "Should include grafana preset")
}
