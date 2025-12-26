package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/ihsanmokhlisse/gitopsi/internal/marketplace"
)

var (
	marketplaceProjectPath string
	marketplaceGitOpsTool  string
	marketplacePlatform    string
)

var marketplaceCmd = &cobra.Command{
	Use:   "marketplace",
	Short: "Browse and install GitOps patterns from the marketplace",
	Long: `Browse, search, and install GitOps patterns from the marketplace.

The marketplace provides curated patterns for common infrastructure needs
like monitoring, logging, security, and networking.

Examples:
  gitopsi marketplace                    # Interactive browser
  gitopsi marketplace search monitoring  # Search for patterns
  gitopsi marketplace info prometheus    # Get pattern details
  gitopsi marketplace categories         # List all categories`,
	RunE: runMarketplaceBrowser,
}

var marketplaceSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for patterns in the marketplace",
	Long: `Search for patterns by name, description, or tags.

Examples:
  gitopsi marketplace search monitoring
  gitopsi marketplace search --category security
  gitopsi marketplace search prometheus --tags alerting`,
	RunE: runMarketplaceSearch,
}

var marketplaceInfoCmd = &cobra.Command{
	Use:   "info [pattern]",
	Short: "Get detailed information about a pattern",
	Long: `Display detailed information about a specific pattern including
components, configuration options, and compatibility.

Examples:
  gitopsi marketplace info prometheus-stack
  gitopsi marketplace info vault-integration`,
	Args: cobra.ExactArgs(1),
	RunE: runMarketplaceInfo,
}

var marketplaceVersionsCmd = &cobra.Command{
	Use:   "versions [pattern]",
	Short: "List available versions of a pattern",
	Args:  cobra.ExactArgs(1),
	RunE:  runMarketplaceVersions,
}

var marketplaceCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List all pattern categories",
	RunE:  runMarketplaceCategories,
}

var marketplaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available patterns",
	RunE:  runMarketplaceList,
}

var (
	searchCategory string
	searchTags     []string
	searchLimit    int
)

func init() {
	rootCmd.AddCommand(marketplaceCmd)

	// Add subcommands
	marketplaceCmd.AddCommand(marketplaceSearchCmd)
	marketplaceCmd.AddCommand(marketplaceInfoCmd)
	marketplaceCmd.AddCommand(marketplaceVersionsCmd)
	marketplaceCmd.AddCommand(marketplaceCategoriesCmd)
	marketplaceCmd.AddCommand(marketplaceListCmd)

	// Common flags
	marketplaceCmd.PersistentFlags().StringVar(&marketplaceProjectPath, "project", ".", "Project path")
	marketplaceCmd.PersistentFlags().StringVar(&marketplaceGitOpsTool, "gitops-tool", "argocd", "GitOps tool (argocd, flux)")
	marketplaceCmd.PersistentFlags().StringVar(&marketplacePlatform, "platform", "kubernetes", "Target platform")

	// Search flags
	marketplaceSearchCmd.Flags().StringVar(&searchCategory, "category", "", "Filter by category")
	marketplaceSearchCmd.Flags().StringSliceVar(&searchTags, "tags", nil, "Filter by tags")
	marketplaceSearchCmd.Flags().IntVar(&searchLimit, "limit", 20, "Maximum results to show")
}

func getMarketplace() *marketplace.Marketplace {
	mp := marketplace.NewMarketplace(marketplaceProjectPath)
	mp.Configure(marketplaceGitOpsTool, marketplacePlatform)
	return mp
}

func runMarketplaceBrowser(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.WithFullWidth().Println("🏪 GitOps Pattern Marketplace")
	fmt.Println()

	// Show categories
	mp := getMarketplace()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	categories, err := mp.ListCategories(ctx)
	if err != nil {
		// Show built-in categories if fetch fails
		categories = getBuiltInCategories()
	}

	pterm.DefaultSection.Println("📦 Categories")

	for _, cat := range categories {
		icon, desc := marketplace.CategoryInfo(cat.Name)
		fmt.Printf("  %s %s (%d patterns)\n", icon, pterm.Bold.Sprint(cat.Name), cat.Count)
		fmt.Printf("     %s\n\n", pterm.FgGray.Sprint(desc))
	}

	// Show popular patterns
	pterm.DefaultSection.Println("⭐ Popular Patterns")

	patterns := marketplace.GetOfficialPatterns()
	for i, p := range patterns {
		if i >= 5 {
			break
		}
		verified := ""
		if p.Verified {
			verified = " ✓"
		}
		fmt.Printf("  • %s%s - %s\n", pterm.Bold.Sprint(p.Name), pterm.FgGreen.Sprint(verified), p.Description)
	}

	fmt.Println()
	pterm.Info.Println("Use 'gitopsi marketplace search <query>' to find patterns")
	pterm.Info.Println("Use 'gitopsi install <pattern>' to install a pattern")

	return nil
}

func runMarketplaceSearch(cmd *cobra.Command, args []string) error {
	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	mp := getMarketplace()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := marketplace.SearchOptions{
		Category: searchCategory,
		Tags:     searchTags,
		Limit:    searchLimit,
	}

	spinner, _ := pterm.DefaultSpinner.Start("Searching marketplace...")

	results, err := mp.Search(ctx, query, opts)
	if err != nil {
		spinner.Fail("Search failed")
		// Fall back to local patterns
		results = searchLocalPatterns(query, opts)
	} else {
		spinner.Success("Search complete")
	}

	if len(results) == 0 {
		pterm.Warning.Printf("No patterns found matching '%s'\n", query)
		return nil
	}

	fmt.Println()
	displayTitle := "Search Results"
	if query != "" {
		displayTitle = fmt.Sprintf("Results for '%s'", query)
	}
	pterm.DefaultSection.Printf("📦 %s (%d patterns)\n", displayTitle, len(results))
	fmt.Println()

	for _, result := range results {
		displayPatternResult(result)
	}

	return nil
}

func displayPatternResult(result marketplace.PatternSearchResult) {
	installed := ""
	if result.Installed {
		installed = pterm.FgGreen.Sprint(" [installed]")
	}

	rating := ""
	if result.Rating > 0 {
		rating = fmt.Sprintf(" ⭐ %.1f", result.Rating)
	}

	fmt.Printf("┌─────────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ %s%s%s\n", pterm.Bold.Sprint(result.Name), installed, rating)
	fmt.Printf("│ %s\n", result.Description)
	if len(result.Tags) > 0 {
		fmt.Printf("│ Tags: %s\n", pterm.FgGray.Sprint(strings.Join(result.Tags, ", ")))
	}
	fmt.Printf("│ Install: %s\n", pterm.FgCyan.Sprintf("gitopsi install %s", result.Name))
	fmt.Printf("└─────────────────────────────────────────────────────────────────┘\n")
	fmt.Println()
}

func runMarketplaceInfo(cmd *cobra.Command, args []string) error {
	patternName := args[0]

	mp := getMarketplace()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Fetching info for '%s'...", patternName))

	info, err := mp.GetPatternInfo(ctx, patternName)
	if err != nil {
		spinner.Fail("Failed to fetch pattern info")
		// Try local info
		info = getLocalPatternInfo(patternName)
		if info == nil {
			return fmt.Errorf("pattern '%s' not found", patternName)
		}
	} else {
		spinner.Success("Info retrieved")
	}

	fmt.Println()
	displayPatternInfo(info)

	return nil
}

func displayPatternInfo(info *marketplace.PatternInfo) {
	pattern := info.Pattern

	// Header
	pterm.DefaultHeader.WithFullWidth().Println(fmt.Sprintf("📦 %s", pattern.Metadata.Name))
	fmt.Println()

	// Basic info
	tableData := [][]string{
		{"Version", pattern.Metadata.Version},
		{"Category", pattern.Metadata.Category},
		{"Author", pattern.Metadata.Author},
		{"License", pattern.Metadata.License},
	}

	if info.Installed {
		tableData = append(tableData, []string{"Status", pterm.FgGreen.Sprint("Installed (v" + info.InstalledVersion + ")")})
	} else {
		tableData = append(tableData, []string{"Status", "Not installed"})
	}

	if info.Verified {
		tableData = append(tableData, []string{"Verified", pterm.FgGreen.Sprint("✓ Yes")})
	}

	if info.Rating > 0 {
		tableData = append(tableData, []string{"Rating", fmt.Sprintf("⭐ %.1f (%d downloads)", info.Rating, info.Downloads)})
	}

	if err := pterm.DefaultTable.WithData(tableData).Render(); err != nil {
		fmt.Printf("Error rendering table: %v\n", err)
	}
	fmt.Println()

	// Description
	pterm.DefaultSection.Println("📝 Description")
	fmt.Println(pattern.Metadata.Description)
	fmt.Println()

	// Tags
	if len(pattern.Metadata.Tags) > 0 {
		pterm.DefaultSection.Println("🏷️  Tags")
		fmt.Println(strings.Join(pattern.Metadata.Tags, ", "))
		fmt.Println()
	}

	// Components
	if len(pattern.Spec.Components) > 0 {
		pterm.DefaultSection.Println("🧩 Components")
		for _, comp := range pattern.Spec.Components {
			fmt.Printf("  • %s (%s)", pterm.Bold.Sprint(comp.Name), comp.Type)
			if comp.Version != "" {
				fmt.Printf(" v%s", comp.Version)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	// Configuration
	if len(pattern.Spec.Config) > 0 {
		pterm.DefaultSection.Println("⚙️  Configuration Options")
		configData := [][]string{{"Key", "Type", "Default", "Description"}}
		for key, item := range pattern.Spec.Config {
			defaultVal := ""
			if item.Default != nil {
				defaultVal = fmt.Sprintf("%v", item.Default)
			}
			required := ""
			if item.Required {
				required = " *"
			}
			configData = append(configData, []string{
				key + required,
				string(item.Type),
				defaultVal,
				item.Description,
			})
		}
		if err := pterm.DefaultTable.WithData(configData).WithHasHeader().Render(); err != nil {
			fmt.Printf("Error rendering config table: %v\n", err)
		}
		fmt.Println()
	}

	// Platform compatibility
	if len(pattern.Spec.Platforms) > 0 {
		pterm.DefaultSection.Println("🖥️  Platform Compatibility")
		for _, p := range pattern.Spec.Platforms {
			version := ""
			if p.MinVersion != "" {
				version = fmt.Sprintf(" (>= %s)", p.MinVersion)
			}
			fmt.Printf("  • %s%s\n", p.Name, version)
		}
		fmt.Println()
	}

	// Dependencies
	if len(pattern.Spec.Dependencies) > 0 {
		pterm.DefaultSection.Println("📦 Dependencies")
		for _, dep := range pattern.Spec.Dependencies {
			optional := ""
			if dep.Optional {
				optional = " (optional)"
			}
			fmt.Printf("  • %s%s\n", dep.Name, optional)
			if dep.Reason != "" {
				fmt.Printf("    %s\n", pterm.FgGray.Sprint(dep.Reason))
			}
		}
		fmt.Println()
	}

	// Available versions
	if len(info.Versions) > 0 {
		pterm.DefaultSection.Println("📋 Available Versions")
		versions := info.Versions
		if len(versions) > 5 {
			versions = versions[:5]
		}
		fmt.Println(strings.Join(versions, ", "))
		if len(info.Versions) > 5 {
			fmt.Printf("... and %d more\n", len(info.Versions)-5)
		}
		fmt.Println()
	}

	// Install command
	pterm.DefaultSection.Println("🚀 Installation")
	fmt.Printf("  %s\n", pterm.FgCyan.Sprintf("gitopsi install %s", pattern.Metadata.Name))
	fmt.Println()
}

func runMarketplaceVersions(cmd *cobra.Command, args []string) error {
	patternName := args[0]

	mp := getMarketplace()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	versions, err := mp.GetRegistry().GetPatternVersions(ctx, patternName)
	if err != nil {
		pterm.Warning.Printf("Could not fetch versions for '%s': %v\n", patternName, err)
		return nil
	}

	pterm.DefaultSection.Printf("📋 Versions of '%s'\n", patternName)
	fmt.Println()

	for _, v := range versions {
		deprecated := ""
		if v.Deprecated {
			deprecated = pterm.FgRed.Sprint(" [deprecated]")
		}
		fmt.Printf("  • %s%s\n", v.Version, deprecated)
		if v.ReleasedAt.Year() > 1 {
			fmt.Printf("    Released: %s\n", v.ReleasedAt.Format("2006-01-02"))
		}
		if v.Changelog != "" {
			fmt.Printf("    %s\n", pterm.FgGray.Sprint(v.Changelog))
		}
	}

	return nil
}

func runMarketplaceCategories(cmd *cobra.Command, args []string) error {
	mp := getMarketplace()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	categories, err := mp.ListCategories(ctx)
	if err != nil {
		// Use built-in categories
		categories = getBuiltInCategories()
	}

	pterm.DefaultSection.Println("📦 Pattern Categories")
	fmt.Println()

	for _, cat := range categories {
		icon, desc := marketplace.CategoryInfo(cat.Name)
		countStr := ""
		if cat.Count > 0 {
			countStr = fmt.Sprintf(" (%d patterns)", cat.Count)
		}
		fmt.Printf("%s %s%s\n", icon, pterm.Bold.Sprint(cat.Name), pterm.FgGray.Sprint(countStr))
		fmt.Printf("   %s\n\n", desc)
	}

	pterm.Info.Println("Use 'gitopsi marketplace search --category <name>' to browse patterns in a category")

	return nil
}

func runMarketplaceList(cmd *cobra.Command, args []string) error {
	mp := getMarketplace()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	spinner, _ := pterm.DefaultSpinner.Start("Fetching patterns...")

	results, err := mp.Search(ctx, "", marketplace.SearchOptions{Limit: 100})
	if err != nil {
		spinner.Warning("Could not fetch remote patterns, showing built-in list")
		results = getBuiltInPatterns()
	} else {
		spinner.Success("Patterns retrieved")
	}

	fmt.Println()
	pterm.DefaultSection.Printf("📦 Available Patterns (%d)\n", len(results))
	fmt.Println()

	// Group by category
	byCategory := make(map[string][]marketplace.PatternSearchResult)
	for _, r := range results {
		cat := r.Category
		if cat == "" {
			cat = "other"
		}
		byCategory[cat] = append(byCategory[cat], r)
	}

	for cat, patterns := range byCategory {
		icon, _ := marketplace.CategoryInfo(cat)
		fmt.Printf("%s %s\n", icon, pterm.Bold.Sprint(strings.Title(cat)))
		for _, p := range patterns {
			installed := ""
			if p.Installed {
				installed = pterm.FgGreen.Sprint(" ✓")
			}
			fmt.Printf("  • %s%s - %s\n", p.Name, installed, pterm.FgGray.Sprint(truncate(p.Description, 50)))
		}
		fmt.Println()
	}

	return nil
}

// Helper functions

func getBuiltInCategories() []marketplace.CategoryIndexEntry {
	categories := marketplace.GetCategories()
	var entries []marketplace.CategoryIndexEntry
	for _, cat := range categories {
		entries = append(entries, marketplace.CategoryIndexEntry{
			Name:        string(cat),
			Description: marketplace.CategoryDescription(cat),
		})
	}
	return entries
}

func getBuiltInPatterns() []marketplace.PatternSearchResult {
	official := marketplace.GetOfficialPatterns()
	var results []marketplace.PatternSearchResult
	for _, p := range official {
		results = append(results, marketplace.PatternSearchResult{
			Name:        p.Name,
			Version:     p.Latest,
			Description: p.Description,
			Category:    p.Category,
			Tags:        p.Tags,
			Rating:      p.Rating,
			Downloads:   p.Downloads,
		})
	}
	return results
}

func searchLocalPatterns(query string, opts marketplace.SearchOptions) []marketplace.PatternSearchResult {
	official := marketplace.GetOfficialPatterns()
	var results []marketplace.PatternSearchResult

	query = strings.ToLower(query)
	for _, p := range official {
		// Filter by category
		if opts.Category != "" && !strings.EqualFold(p.Category, opts.Category) {
			continue
		}

		// Match query
		if query != "" {
			if !strings.Contains(strings.ToLower(p.Name), query) &&
				!strings.Contains(strings.ToLower(p.Description), query) {
				continue
			}
		}

		results = append(results, marketplace.PatternSearchResult{
			Name:        p.Name,
			Version:     p.Latest,
			Description: p.Description,
			Category:    p.Category,
			Tags:        p.Tags,
			Rating:      p.Rating,
			Downloads:   p.Downloads,
		})
	}

	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results
}

func getLocalPatternInfo(name string) *marketplace.PatternInfo {
	official := marketplace.GetOfficialPatterns()
	for _, p := range official {
		if p.Name == name {
			return &marketplace.PatternInfo{
				Pattern: marketplace.Pattern{
					APIVersion: "gitopsi.io/v1",
					Kind:       "Pattern",
					Metadata: marketplace.PatternMetadata{
						Name:        p.Name,
						Version:     p.Latest,
						Description: p.Description,
						Category:    p.Category,
						Tags:        p.Tags,
						License:     "MIT",
					},
				},
				Versions:  p.Versions,
				Rating:    p.Rating,
				Downloads: p.Downloads,
				Verified:  p.Verified,
			}
		}
	}
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Install command
var installCmd = &cobra.Command{
	Use:   "install [pattern]",
	Short: "Install a GitOps pattern from the marketplace",
	Long: `Install a pattern from the marketplace into your GitOps repository.

Examples:
  gitopsi install prometheus-stack
  gitopsi install prometheus-stack --version 1.2.0
  gitopsi install prometheus-stack --config values.yaml
  gitopsi install prometheus-stack --env dev,staging
  gitopsi install prometheus-stack --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runInstall,
}

var patternsCmd = &cobra.Command{
	Use:   "patterns",
	Short: "Manage installed patterns",
	Long:  `List, update, and remove installed patterns.`,
}

var patternsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed patterns",
	RunE:  runPatternsList,
}

var patternsUpdateCmd = &cobra.Command{
	Use:   "update [pattern]",
	Short: "Update an installed pattern",
	Args:  cobra.ExactArgs(1),
	RunE:  runPatternsUpdate,
}

var patternsRemoveCmd = &cobra.Command{
	Use:   "remove [pattern]",
	Short: "Remove an installed pattern",
	Args:  cobra.ExactArgs(1),
	RunE:  runPatternsRemove,
}

var patternsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of installed patterns",
	RunE:  runPatternsStatus,
}

var patternCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new pattern scaffold",
	Args:  cobra.ExactArgs(1),
	RunE:  runPatternCreate,
}

var (
	installVersion  string
	installConfig   string
	installEnvs     []string
	installDryRun   bool
	installForce    bool
	installSkipDeps bool
	patternCategory string
)

func init() {
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(patternsCmd)
	rootCmd.AddCommand(patternCreateCmd)

	patternsCmd.AddCommand(patternsListCmd)
	patternsCmd.AddCommand(patternsUpdateCmd)
	patternsCmd.AddCommand(patternsRemoveCmd)
	patternsCmd.AddCommand(patternsStatusCmd)

	// Install flags
	installCmd.Flags().StringVar(&installVersion, "version", "", "Pattern version to install")
	installCmd.Flags().StringVar(&installConfig, "config", "", "Path to configuration file")
	installCmd.Flags().StringSliceVar(&installEnvs, "env", nil, "Target environments")
	installCmd.Flags().BoolVar(&installDryRun, "dry-run", false, "Preview changes without applying")
	installCmd.Flags().BoolVar(&installForce, "force", false, "Force reinstall if already installed")
	installCmd.Flags().BoolVar(&installSkipDeps, "skip-deps", false, "Skip dependency installation")

	// Pattern create flags
	patternCreateCmd.Flags().StringVar(&patternCategory, "category", "infrastructure", "Pattern category")
}

func runInstall(cmd *cobra.Command, args []string) error {
	patternName := args[0]

	mp := getMarketplace()
	ctx := context.Background()

	// Load config if provided
	var config map[string]any
	if installConfig != "" {
		// TODO: Load config from file
		config = make(map[string]any)
	}

	opts := marketplace.InstallOptions{
		Version:      installVersion,
		Config:       config,
		Environments: installEnvs,
		DryRun:       installDryRun,
		Force:        installForce,
		SkipDeps:     installSkipDeps,
	}

	if installDryRun {
		pterm.Info.Println("Dry run mode - no changes will be made")
	}

	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Installing %s...", patternName))

	result, err := mp.Install(ctx, patternName, opts)
	if err != nil {
		spinner.Fail("Installation failed")
		return err
	}

	if result.Success {
		spinner.Success(result.Message)
	} else {
		spinner.Warning(result.Message)
	}

	// Show results
	if len(result.GeneratedPath) > 0 {
		fmt.Println()
		pterm.DefaultSection.Println("📁 Generated Files")
		for _, path := range result.GeneratedPath {
			fmt.Printf("  • %s\n", path)
		}
	}

	if len(result.Dependencies) > 0 {
		fmt.Println()
		pterm.DefaultSection.Println("📦 Dependencies")
		for _, dep := range result.Dependencies {
			status := pterm.FgGreen.Sprint("✓")
			if dep.Status == "failed" {
				status = pterm.FgRed.Sprint("✗")
			} else if dep.Status == "skipped" {
				status = pterm.FgYellow.Sprint("○")
			}
			fmt.Printf("  %s %s - %s\n", status, dep.Name, dep.Message)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println()
		for _, warning := range result.Warnings {
			pterm.Warning.Println(warning)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Println()
		for _, err := range result.Errors {
			pterm.Error.Println(err)
		}
	}

	if result.Success && !installDryRun {
		fmt.Println()
		pterm.Success.Println("Pattern installed successfully!")
		pterm.Info.Println("Commit and push your changes to apply the pattern")
	}

	return nil
}

func runPatternsList(cmd *cobra.Command, args []string) error {
	mp := getMarketplace()

	patterns, err := mp.ListInstalled()
	if err != nil {
		return err
	}

	if len(patterns) == 0 {
		pterm.Info.Println("No patterns installed")
		pterm.Info.Println("Use 'gitopsi marketplace search' to find patterns")
		return nil
	}

	pterm.DefaultSection.Printf("📦 Installed Patterns (%d)\n", len(patterns))
	fmt.Println()

	data := [][]string{{"Pattern", "Version", "Status", "Installed"}}
	for _, p := range patterns {
		data = append(data, []string{
			p.Pattern.Metadata.Name,
			p.Pattern.Metadata.Version,
			p.Status,
			p.InstalledAt.Format("2006-01-02"),
		})
	}

	if err := pterm.DefaultTable.WithData(data).WithHasHeader().Render(); err != nil {
		return err
	}

	return nil
}

func runPatternsUpdate(cmd *cobra.Command, args []string) error {
	patternName := args[0]

	mp := getMarketplace()
	ctx := context.Background()

	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Updating %s...", patternName))

	result, err := mp.Update(ctx, patternName, marketplace.UpdateOptions{
		Force: installForce,
	})
	if err != nil {
		spinner.Fail("Update failed")
		return err
	}

	if result.Success {
		spinner.Success(result.Message)
	} else {
		spinner.Warning(result.Message)
	}

	return nil
}

func runPatternsRemove(cmd *cobra.Command, args []string) error {
	patternName := args[0]

	mp := getMarketplace()
	ctx := context.Background()

	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Removing %s...", patternName))

	err := mp.Uninstall(ctx, patternName, marketplace.UninstallOptions{
		Force: installForce,
	})
	if err != nil {
		spinner.Fail("Removal failed")
		return err
	}

	spinner.Success(fmt.Sprintf("Pattern '%s' removed successfully", patternName))
	return nil
}

func runPatternsStatus(cmd *cobra.Command, args []string) error {
	mp := getMarketplace()
	ctx := context.Background()

	status, err := mp.GetStatus(ctx)
	if err != nil {
		return err
	}

	if len(status) == 0 {
		pterm.Info.Println("No patterns installed")
		return nil
	}

	pterm.DefaultSection.Println("📊 Pattern Status")
	fmt.Println()

	for name, state := range status {
		icon := pterm.FgGreen.Sprint("✓")
		if state != "healthy" {
			icon = pterm.FgRed.Sprint("✗")
		}
		fmt.Printf("  %s %s: %s\n", icon, name, state)
	}

	// Check for updates
	fmt.Println()
	spinner, _ := pterm.DefaultSpinner.Start("Checking for updates...")

	updates, err := mp.CheckUpdates(ctx)
	if err != nil {
		spinner.Warning("Could not check for updates")
	} else if len(updates) > 0 {
		spinner.Success("Updates available")
		fmt.Println()
		pterm.DefaultSection.Println("🔄 Available Updates")
		for name, version := range updates {
			fmt.Printf("  • %s → %s\n", name, version)
		}
		pterm.Info.Println("\nUse 'gitopsi patterns update <pattern>' to update")
	} else {
		spinner.Success("All patterns are up to date")
	}

	return nil
}

func runPatternCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	mp := getMarketplace()

	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Creating pattern '%s'...", name))

	pattern, err := mp.CreatePattern(name, patternCategory)
	if err != nil {
		spinner.Fail("Failed to create pattern")
		return err
	}

	spinner.Success("Pattern created")
	fmt.Println()

	pterm.DefaultSection.Printf("📦 Pattern '%s' created\n", name)
	fmt.Println()

	fmt.Printf("  Category: %s\n", pattern.Metadata.Category)
	fmt.Printf("  Version:  %s\n", pattern.Metadata.Version)
	fmt.Printf("  Path:     %s/\n", name)
	fmt.Println()

	pterm.Info.Println("Files created:")
	fmt.Printf("  • %s/pattern.yaml\n", name)
	fmt.Printf("  • %s/README.md\n", name)
	fmt.Println()

	pterm.Info.Println("Next steps:")
	fmt.Println("  1. Edit pattern.yaml to define components")
	fmt.Println("  2. Add configuration options")
	fmt.Println("  3. Test with 'gitopsi pattern validate'")
	fmt.Println("  4. Share with 'gitopsi pattern publish'")

	return nil
}

// Pattern validate command
var patternValidateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate a pattern",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runPatternValidate,
}

func init() {
	rootCmd.AddCommand(patternValidateCmd)
}

func runPatternValidate(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	// Check if pattern.yaml exists
	patternFile := path
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		patternFile = path
	}

	mp := marketplace.NewMarketplace(path)

	spinner, _ := pterm.DefaultSpinner.Start("Validating pattern...")

	errors, err := mp.ValidatePattern(patternFile)
	if err != nil {
		spinner.Fail("Validation failed")
		return err
	}

	if len(errors) > 0 {
		spinner.Warning("Validation completed with warnings")
		fmt.Println()
		for _, e := range errors {
			pterm.Warning.Println(e)
		}
	} else {
		spinner.Success("Pattern is valid")
	}

	return nil
}
