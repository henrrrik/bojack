package view

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/henrrrik/runway-tui/internal/msg"
)

// Typed values become CLI arguments, so they must not start with "-".
var (
	configKeyRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	planRE      = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
)

var errConfigKey = errors.New("config keys are letters, digits and _, not starting with a digit")

// checkAssignment accepts KEY=value and shows only KEY.
func checkAssignment(s string) (string, error) {
	k, _, ok := strings.Cut(s, "=")
	if !ok {
		return "", errors.New("use KEY=value")
	}
	if !configKeyRE.MatchString(k) {
		return "", errConfigKey
	}
	return k, nil
}

func checkConfigKey(s string) (string, error) {
	if !configKeyRE.MatchString(s) {
		return "", errConfigKey
	}
	return s, nil
}

func checkPlan(s string) (string, error) {
	if !planRE.MatchString(s) {
		return "", errors.New("enter a plan name, e.g. starter (see Plans in the main menu)")
	}
	return s, nil
}

// echoAssignment masks the value part of KEY=value while it is typed, by
// the same rules as the Config tab.
func echoAssignment(s string) string {
	k, v, ok := strings.Cut(s, "=")
	if !ok {
		return s
	}
	return k + "=" + maskConfigValue(k, v)
}

func setConfigAction(kind msg.ActionKind, noun string) actionSpec {
	return actionSpec{kind: kind, key: "S", help: "S: set", input: "KEY=value",
		question: "Set %[2]s on " + noun + " %[1]s?", running: "Setting %[2]s…", done: "Set %[2]s on %[1]s.",
		failed: "Setting config failed", check: checkAssignment, echo: echoAssignment}
}

func removeConfigAction(kind msg.ActionKind, noun string) actionSpec {
	return actionSpec{kind: kind, key: "D", help: "D: remove", input: "Key to remove",
		question: "Remove %[2]s from " + noun + " %[1]s?", running: "Removing %[2]s…",
		done: "Removed %[2]s from %[1]s.", failed: "Removing config failed", check: checkConfigKey}
}

var (
	appConfigActions = []actionSpec{
		setConfigAction(msg.ActionSetAppConfig, "app"),
		removeConfigAction(msg.ActionRemoveAppConfig, "app"),
	}
	serviceConfigActions = []actionSpec{
		setConfigAction(msg.ActionSetServiceConfig, "service"),
		removeConfigAction(msg.ActionRemoveServiceConfig, "service"),
	}
	appPlanAction = actionSpec{kind: msg.ActionSetAppPlan, key: "P", help: "P: plan", input: "New plan",
		question: "Move app %[1]s to plan %[2]s? This changes what you pay.", running: "Moving %[1]s to %[2]s…",
		done: "Moved %[1]s to plan %[2]s.", failed: "Plan change failed", check: checkPlan}
)

// resizeVolumeAction grows a volume that is currently sizeMB.
func resizeVolumeAction(sizeMB int) actionSpec {
	cur := formatMB(strconv.Itoa(sizeMB))
	return actionSpec{kind: msg.ActionResizeVolume, key: "S", help: "S: resize", input: "New size in MiB",
		question: "Grow volume %[1]s from " + cur + " to %[2]s? Volumes cannot shrink.",
		running:  "Resizing %[1]s…", done: "Resized %[1]s to %[2]s.", failed: "Resize failed",
		check: func(s string) (string, error) {
			n, err := strconv.Atoi(s)
			if err != nil || n <= sizeMB {
				return "", fmt.Errorf("enter a whole number of MiB, more than %d", sizeMB)
			}
			return formatMB(s), nil
		}}
}
