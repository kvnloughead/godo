package cmd

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
)

// unarchiveCmd marks a todo item as not archived.
var unarchiveCmd = &cobra.Command{
	Use:   "unarchive <ID>",
	Short: "Mark a todo item as not archived",
	Long: `
Mark a todo item as not archived. For example:

    # Mark todo number 42 as not archived
    godo unarchive 42

This command requires authentication. Run 'godo auth -h' for more information.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil || id < 1 {
			fmt.Println("Error: ID must be a positive integer")
			return
		}

		url := fmt.Sprintf("%s/todos/%d", app.Config.APIBaseURL, id)
		stdoutMsg := "\nError: failed to mark todo as not archived. \nCheck `~/.config/godo/logs` for details.\n"

		handleError := func(logMsg string, err error) {
			app.handleError(logMsg, stdoutMsg, err,
				"method", http.MethodPatch,
				"url", url)
		}

		token, err := app.TokenManager.LoadToken()
		if err != nil {
			app.handleAuthenticationError("Failed to read token", err)
			return
		}

		// Create the payload with completed = false
		payload := map[string]any{"archived": false}

		req, err := app.createJSONRequest(http.MethodPatch, url, payload)
		if err != nil {
			handleError("Failed to create request", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+string(token))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			handleError("Failed to send request", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			switch resp.StatusCode {
			case http.StatusNotFound:
				fmt.Println("Error: todo not found")
			default:
				handleError("Failed to mark todo as not completed", fmt.Errorf("response status: %s", resp.Status))
			}
			return
		}

		fmt.Println("Todo marked as not archived")
	},
}

func init() {
	rootCmd.AddCommand(unarchiveCmd)
}
