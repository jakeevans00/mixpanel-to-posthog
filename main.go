package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/manifoldco/promptui"
	"github.com/posthog/posthog-go"
)

var version = "dev"

// Delay between posthog queue events to avoid overloading the API
const DELAY_MS = 1

func getPosthogClient() posthog.Client {
	// ** Get Posthog credentials ** //
	if os.Getenv("POSTHOG_API_KEY") == "" || os.Getenv("POSTHOG_ENDPOINT") == "" || os.Getenv("POSTHOG_PROJECT_KEY") == "" {
		color.Cyan("\nPosthog Credentials")
		color.Cyan("See the README for reference on what these are and how to get them.\n\n")
	}

	// If in env, don't ask
	var posthogApiKey string
	if os.Getenv("POSTHOG_PROJECT_KEY") != "" {
		posthogApiKey = os.Getenv("POSTHOG_PROJECT_KEY")
	} else {
		posthogApiKeyPrompt := promptui.Prompt{
			Label: "Enter Posthog Project API Key",
			Mask:  '*',
		}
		pR, _ := posthogApiKeyPrompt.Run()
		posthogApiKey = pR
	}

	var posthogPersonalApiKey string
	if os.Getenv("POSTHOG_API_KEY") != "" {
		posthogPersonalApiKey = os.Getenv("POSTHOG_API_KEY")
	} else {
		posthogApiKeyPrompt := promptui.Prompt{
			Label: "Enter Posthog Personal API Key",
			Mask:  '*',
		}
		pR, _ := posthogApiKeyPrompt.Run()
		posthogPersonalApiKey = pR
	}

	// If in env, don't ask
	var posthogEndpoint string
	if os.Getenv("POSTHOG_ENDPOINT") != "" {
		posthogEndpoint = os.Getenv("POSTHOG_ENDPOINT")
	} else {
		posthogApiKeyPrompt := promptui.Prompt{
			Label: "Enter Posthog API Endpoint",
			Validate: func(input string) error {
				_, err := url.Parse(input)
				return err
			},
		}
		pR, _ := posthogApiKeyPrompt.Run()
		posthogEndpoint = pR
	}

	// Create posthog client
	posthogClient, err := posthog.NewWithConfig(posthogApiKey, posthog.Config{
		Endpoint:       posthogEndpoint,
		PersonalApiKey: posthogPersonalApiKey,
		HistoricalMigration: true,
	})
	if err != nil {
		color.Red("\nEncountered an error while creating Posthog client: %v", err)
		os.Exit(1)
	}
	return posthogClient
}

func getChunkSize() int {
	prompt := promptui.Prompt{
		Label:     "Enter chunk size in days",
		Default:   "7",
		AllowEdit: true,
	}
	result, err := prompt.Run()
	if err != nil {
		log.Fatal(err)
	}
	var chunkSize int
	fmt.Sscanf(result, "%d", &chunkSize)
	if chunkSize <= 0 {
		log.Fatal("Chunk size must be greater than 0")
	}
	return chunkSize
}

func chunkDateRange(fromDate, toDate time.Time, chunkSizeDays int) [][2]time.Time {
	chunks := [][2]time.Time{}
	currentDate := fromDate

	for currentDate.Before(toDate) || currentDate.Equal(toDate) {
		chunkEnd := currentDate.AddDate(0, 0, chunkSizeDays)
		if chunkEnd.After(toDate) {
			chunkEnd = toDate
		}
		chunks = append(chunks, [2]time.Time{currentDate, chunkEnd})
		currentDate = chunkEnd.AddDate(0, 0, 1)
	}

	return chunks
}

func main() {
	godotenv.Load(".env")

	fmt.Println("------------------------------------")
	color.Green("SC Mixpanel to Posthog Data Migrator")
	fmt.Println("------------------------------------")

	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("\nVersion: %v\n", color.GreenString(version))
		os.Exit(0)
	}


	// ** Get mixpanel credentials ** //

	if os.Getenv("MIXPANEL_API_URL") == "" || os.Getenv("MIXPANEL_PROJECT_ID") == "" || os.Getenv("MIXPANEL_USERNAME") == "" || os.Getenv("MIXPANEL_PASSWORD") == "" {
		color.Cyan("\nMixpanel Credentials")
		color.Cyan("See the README for reference on what these are and how to get them.\n\n")
	}
	// If in env, don't ask
	var apiUrlResult string
	if os.Getenv("MIXPANEL_API_URL") != "" {
		apiUrlResult = os.Getenv("MIXPANEL_API_URL")
	} else {
		apiUrlPrompt := promptui.Prompt{
			Label:     "Enter Mixpanel API URL (for EU, use the EU-specific URL):",
			AllowEdit: false,
			Default:   "https://data.mixpanel.com/api/2.0",
			Validate: func(input string) error {
				// Validate URL
				_, err := url.ParseRequestURI(input)
				return err
			},
		}
		pR, _ := apiUrlPrompt.Run()
		apiUrlResult = pR
	}

	// If in env, don't ask
	var projectIdResult string
	if os.Getenv("MIXPANEL_PROJECT_ID") != "" {
		projectIdResult = os.Getenv("MIXPANEL_PROJECT_ID")
	} else {
		projectIdPrompt := promptui.Prompt{
			Label: "Enter Mixpanel Project ID",
		}
		pR, _ := projectIdPrompt.Run()
		projectIdResult = pR
	}

	// If in env, don't ask
	var serviceUsernameResult string
	if os.Getenv("MIXPANEL_USERNAME") != "" {
		serviceUsernameResult = os.Getenv("MIXPANEL_USERNAME")
	} else {
		serviceUsernamePrompt := promptui.Prompt{
			Label: "Enter Mixpanel Username (Service Account)",
		}
		pR, _ := serviceUsernamePrompt.Run()
		serviceUsernameResult = pR
	}

	// If in env, don't ask
	var servicePasswordResult string
	if os.Getenv("MIXPANEL_PASSWORD") != "" {
		servicePasswordResult = os.Getenv("MIXPANEL_PASSWORD")
	} else {
		servicePasswordPrompt := promptui.Prompt{
			Label: "Enter Mixpanel Password (Service Account)",
			Mask:  '*',
		}
		pR, _ := servicePasswordPrompt.Run()
		servicePasswordResult = pR
	}

	// ** Get Mixpanel date range ** //

	color.Yellow("\nWARNING: If you have a large dataset, consider entering smaller date ranges at a time.")
	color.Yellow("You may crash your machine if you try to export too much data at once.\n\n")

	// Get date range
	fromDtPrompt := promptui.Prompt{
		Label: "Enter from_date in the format YYYY-MM-DD",
		Validate: func(input string) error {
			_, err := time.Parse("2006-01-02", input)
			return err
		},
	}
	fromDtResult, err := fromDtPrompt.Run()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	toDtPrompt := promptui.Prompt{
		Label: "Enter to_date in the format YYYY-MM-DD",
		Validate: func(input string) error {
			_, err := time.Parse("2006-01-02", input)
			return err
		},
	}
	toDtResult, err := toDtPrompt.Run()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Parse dates
	fromDt, _ := time.Parse("2006-01-02", fromDtResult)
	toDt, _ := time.Parse("2006-01-02", toDtResult)

	// Get chunk size
	chunkSize := getChunkSize()

	// Split date range into chunks
	dateChunks := chunkDateRange(fromDt, toDt, chunkSize)
	totalChunks := len(dateChunks)

	color.Cyan("\nProcessing %d chunks of %d days each", totalChunks, chunkSize)
	color.Cyan("Total date range: %s to %s", fromDt.Format("2006-01-02"), toDt.Format("2006-01-02"))

	posthogClient := getPosthogClient()
	defer posthogClient.Close()

	s := spinner.New(spinner.CharSets[43], 100*time.Millisecond)
	s.Start()
	startTimestamp := time.Now()

	totalEventsImported := 0

	for i, chunk := range dateChunks {
		chunkStart, chunkEnd := chunk[0], chunk[1]
		s.Suffix = fmt.Sprintf(" Processing %s to %s", chunkStart.Format("2006-01-02"), chunkEnd.Format("2006-01-02"))

		// Create mixpanel exporter for this chunk
		exporter := NewExporter(version, apiUrlResult, serviceUsernameResult, servicePasswordResult, projectIdResult, chunkStart, chunkEnd)
		data, err := exporter.Export()
		if err != nil {
			color.Red("\nEncountered an error while exporting data from Mixpanel: %v", err)
			os.Exit(1)
		}

		// Log the number of records exported
		color.Cyan("Exported %d records from Mixpanel for %s to %s", len(data), chunkStart.Format("2006-01-02"), chunkEnd.Format("2006-01-02"))

		// Import to PostHog
		importedCount, err := PosthogImport(posthogClient, data)
		if err != nil {
			color.Red("\nEncountered an error while importing data into Posthog: %v", err)
			os.Exit(1)
		}

		totalEventsImported += importedCount
		color.Green("Imported %d events to PostHog for chunk %d/%d", importedCount, i+1, totalChunks)
		color.Green("Completed chunk %d/%d", i+1, totalChunks)
	}

	s.Stop()
	color.Green("\nSuccess! All chunks processed.")
	color.Green("Total events imported: %d", totalEventsImported)
	color.Green("Total time taken: %s", time.Since(startTimestamp))
	// Block until user presses control C
	color.Red("It's recommended to wait several minutes for posthog to process the events.")
	color.Red("Once you see all events in posthog, you can exit this program.")
	color.Red("Press control C to exit...")
	select {}
}
