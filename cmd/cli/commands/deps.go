package commands

import (
	"github.com/YubiApp/cmd/cli/utils"
	"github.com/YubiApp/internal/config"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

// Global dependencies
var (
	DB *gorm.DB
	Cfg *config.Config
)

// SetDependencies sets the global dependencies for all command packages
func SetDependencies(db *gorm.DB, cfg *config.Config) {
	DB = db
	Cfg = cfg
}

// InitMigrationCommand initializes the migration command
func InitMigrationCommand() *cobra.Command {
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Long:  "Run database migrations to ensure the database schema is up to date",
		RunE: func(cmd *cobra.Command, args []string) error {
			return utils.RunMigrations(DB)
		},
	}
	
	return migrateCmd
} 