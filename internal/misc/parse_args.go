package misc

import (
	"fmt"

	"os"

	"github.com/spf13/cobra"
)

func exitIfError(flag string, err error) {
	if err != nil {
		fmt.Printf("cannot parse %s: %v\n", flag, err)
		os.Exit(1)
	}

}

func ParseOrExit[T ~string | ~int | ~[]string](cmd *cobra.Command, flag string) T {
	got := parseOrExitInternal[T](cmd, flag)
	return got.(T)
}

func parseOrExitInternal[T ~string | ~int | ~[]string](cmd *cobra.Command, flag string) interface{} {
	var value T
	switch v := any(value).(type) {
	case string:
		v, err := cmd.Flags().GetString(flag)
		exitIfError(flag, err)
		return v
	case int:
		v, err := cmd.Flags().GetInt(flag)
		exitIfError(flag, err)
		return v
	case []string:
		v, err := cmd.Flags().GetStringSlice(flag)
		exitIfError(flag, err)
		return v
	default:
		fmt.Printf("cannot parse %s: unknown type %T\n", flag, v)
		os.Exit(1)
	}

	return nil
}
