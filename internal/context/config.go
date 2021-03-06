package context

import (
	"log"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/photoprism/photoprism/internal/frontend"
	"github.com/photoprism/photoprism/internal/fsutil"
	"github.com/photoprism/photoprism/internal/models"
	"github.com/photoprism/photoprism/internal/tidb"
	"github.com/urfave/cli"
)

const (
	DbTiDB  = "internal"
	DbMySQL = "mysql"
)

// Config provides a struct in which application configuration is stored.
// Application code must use functions to get config values, for two reasons:
//
// 1. Some values are computed and we don't want to leak implementation details (aims at reducing refactoring overhead).
//
// 2. Paths might actually be dynamic later (if we build a multi-user version).
//
// See https://github.com/photoprism/photoprism/issues/50#issuecomment-433856358
type Config struct {
	appName            string
	appVersion         string
	appCopyright       string
	debug              bool
	configFile         string
	assetsPath         string
	cachePath          string
	originalsPath      string
	importPath         string
	exportPath         string
	sqlServerHost      string
	sqlServerPort      uint
	sqlServerPath      string
	sqlServerPassword  string
	httpServerHost     string
	httpServerPort     int
	httpServerMode     string
	httpServerPassword string
	darktableCli       string
	databaseDriver     string
	databaseDsn        string
	db                 *gorm.DB
}

// NewConfig() creates a new configuration entity by using two methods:
//
// 1. SetValuesFromFile: This will initialize values from a yaml config file.
//
// 2. SetValuesFromCliContext: Which comes after SetValuesFromFile and overrides
//    any previous values giving an option two override file configs through the CLI.
func NewConfig(ctx *cli.Context) *Config {
	c := &Config{}
	c.appName = ctx.App.Name
	c.appCopyright = ctx.App.Copyright
	c.appVersion = ctx.App.Version
	c.SetValuesFromFile(fsutil.ExpandedFilename(ctx.GlobalString("config-file")))
	c.SetValuesFromCliContext(ctx)

	return c
}

// SetValuesFromFile uses a yaml config file to initiate the configuration entity.
func (c *Config) SetValuesFromFile(fileName string) error {
	yamlConfig, err := yaml.ReadFile(fileName)

	if err != nil {
		return err
	}

	c.configFile = fileName
	if debug, err := yamlConfig.GetBool("debug"); err == nil {
		c.debug = debug
	}

	if sqlServerHost, err := yamlConfig.Get("sql-host"); err == nil {
		c.sqlServerHost = sqlServerHost
	}

	if sqlServerPort, err := yamlConfig.GetInt("sql-port"); err == nil {
		c.sqlServerPort = uint(sqlServerPort)
	}

	if sqlServerPassword, err := yamlConfig.Get("sql-password"); err == nil {
		c.sqlServerPassword = sqlServerPassword
	}

	if sqlServerPath, err := yamlConfig.Get("sql-path"); err == nil {
		c.sqlServerPath = sqlServerPath
	}

	if httpServerHost, err := yamlConfig.Get("http-host"); err == nil {
		c.httpServerHost = httpServerHost
	}

	if httpServerPort, err := yamlConfig.GetInt("http-port"); err == nil {
		c.httpServerPort = int(httpServerPort)
	}

	if httpServerMode, err := yamlConfig.Get("http-mode"); err == nil {
		c.httpServerMode = httpServerMode
	}

	if httpServerPassword, err := yamlConfig.Get("http-password"); err == nil {
		c.httpServerPassword = httpServerPassword
	}

	if assetsPath, err := yamlConfig.Get("assets-path"); err == nil {
		c.assetsPath = fsutil.ExpandedFilename(assetsPath)
	}

	if cachePath, err := yamlConfig.Get("cache-path"); err == nil {
		c.cachePath = fsutil.ExpandedFilename(cachePath)
	}

	if originalsPath, err := yamlConfig.Get("originals-path"); err == nil {
		c.originalsPath = fsutil.ExpandedFilename(originalsPath)
	}

	if importPath, err := yamlConfig.Get("import-path"); err == nil {
		c.importPath = fsutil.ExpandedFilename(importPath)
	}

	if exportPath, err := yamlConfig.Get("export-path"); err == nil {
		c.exportPath = fsutil.ExpandedFilename(exportPath)
	}

	if darktableCli, err := yamlConfig.Get("darktable-cli"); err == nil {
		c.darktableCli = fsutil.ExpandedFilename(darktableCli)
	}

	if databaseDriver, err := yamlConfig.Get("database-driver"); err == nil {
		c.databaseDriver = databaseDriver
	}

	if databaseDsn, err := yamlConfig.Get("database-dsn"); err == nil {
		c.databaseDsn = databaseDsn
	}

	return nil
}

// SetValuesFromCliContext uses values from the CLI to setup configuration overrides
// for the entity.
func (c *Config) SetValuesFromCliContext(ctx *cli.Context) error {
	if ctx.GlobalBool("debug") {
		c.debug = ctx.GlobalBool("debug")
	}

	if ctx.GlobalIsSet("assets-path") || c.assetsPath == "" {
		c.assetsPath = fsutil.ExpandedFilename(ctx.GlobalString("assets-path"))
	}

	if ctx.GlobalIsSet("cache-path") || c.cachePath == "" {
		c.cachePath = fsutil.ExpandedFilename(ctx.GlobalString("cache-path"))
	}

	if ctx.GlobalIsSet("originals-path") || c.originalsPath == "" {
		c.originalsPath = fsutil.ExpandedFilename(ctx.GlobalString("originals-path"))
	}

	if ctx.GlobalIsSet("import-path") || c.importPath == "" {
		c.importPath = fsutil.ExpandedFilename(ctx.GlobalString("import-path"))
	}

	if ctx.GlobalIsSet("export-path") || c.exportPath == "" {
		c.exportPath = fsutil.ExpandedFilename(ctx.GlobalString("export-path"))
	}

	if ctx.GlobalIsSet("darktable-cli") || c.darktableCli == "" {
		c.darktableCli = fsutil.ExpandedFilename(ctx.GlobalString("darktable-cli"))
	}

	if ctx.GlobalIsSet("database-driver") || c.databaseDriver == "" {
		c.databaseDriver = ctx.GlobalString("database-driver")
	}

	if ctx.GlobalIsSet("database-dsn") || c.databaseDsn == "" {
		c.databaseDsn = ctx.GlobalString("database-dsn")
	}

	if ctx.GlobalIsSet("sql-host") || c.sqlServerHost == "" {
		c.sqlServerHost = ctx.GlobalString("sql-host")
	}

	if ctx.GlobalIsSet("sql-port") || c.sqlServerPort == 0 {
		c.sqlServerPort = ctx.GlobalUint("sql-port")
	}

	if ctx.GlobalIsSet("sql-path") || c.sqlServerPath == "" {
		c.sqlServerPath = ctx.GlobalString("sql-path")
	}

	if ctx.GlobalIsSet("http-host") || c.httpServerHost == "" {
		c.httpServerHost = ctx.GlobalString("http-host")
	}

	if ctx.GlobalIsSet("http-port") || c.httpServerPort == 0 {
		c.httpServerPort = ctx.GlobalInt("http-port")
	}

	if ctx.GlobalIsSet("http-mode") || c.httpServerMode == "" {
		c.httpServerMode = ctx.GlobalString("http-mode")
	}

	return nil
}

// CreateDirectories creates all the folders that photoprism needs. These are:
// originalsPath
// ThumbnailsPath
// importPath
// exportPath
func (c *Config) CreateDirectories() error {
	if err := os.MkdirAll(c.OriginalsPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.ImportPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.ExportPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.ThumbnailsPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.SqlServerPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.TensorFlowModelPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.HttpPublicBuildPath(), os.ModePerm); err != nil {
		return err
	}

	return nil
}

// connectToDatabase establishes a database connection.
// When used with the internal driver, it may create a new database server instance.
// It tries to do this 12 times with a 5 second sleep interval in between.
func (c *Config) connectToDatabase() error {
	dbDriver := c.DatabaseDriver()
	dbDsn := c.DatabaseDsn()

	isTiDB := false
	initSuccess := false

	if dbDriver == DbTiDB {
		isTiDB = true
		dbDriver = DbMySQL
	}

	db, err := gorm.Open(dbDriver, dbDsn)

	if err != nil || db == nil {
		if isTiDB {
			go tidb.Start(c.SqlServerPath(), c.SqlServerPort(), c.SqlServerHost(), c.Debug())
		}

		for i := 1; i <= 12; i++ {
			time.Sleep(5 * time.Second)

			db, err = gorm.Open(dbDriver, dbDsn)

			if db != nil && err == nil {
				break
			}

			if isTiDB && !initSuccess {
				err = tidb.InitDatabase(c.SqlServerPort())

				if err != nil {
					log.Println(err)
				} else {
					initSuccess = true
				}
			}
		}

		if err != nil || db == nil {
			log.Fatal(err)
		}
	}

	c.db = db

	return err
}

// AppName returns the application name.
func (c *Config) AppName() string {
	return c.appName
}

// AppVersion returns the application version.
func (c *Config) AppVersion() string {
	return c.appVersion
}

// AppCopyright returns the application copyright.
func (c *Config) AppCopyright() string {
	return c.appCopyright
}

// Debug returns true if debug mode is on.
func (c *Config) Debug() bool {
	return c.debug
}

// ConfigFile returns the config file name.
func (c *Config) ConfigFile() string {
	return c.configFile
}

// SqlServerHost returns the built-in SQL server host name or IP address (empty for all interfaces).
func (c *Config) SqlServerHost() string {
	return c.sqlServerHost
}

// SqlServerPort returns the built-in SQL server port.
func (c *Config) SqlServerPort() uint {
	return c.sqlServerPort
}

// SqlServerPath returns the database storage path for TiDB.
func (c *Config) SqlServerPath() string {
	if c.sqlServerPath != "" {
		return c.sqlServerPath
	}

	return c.ServerPath() + "/database"
}

// SqlServerPassword returns the password for the built-in database server.
func (c *Config) SqlServerPassword() string {
	return c.sqlServerPassword
}

// HttpServerHost returns the built-in HTTP server host name or IP address (empty for all interfaces).
func (c *Config) HttpServerHost() string {
	return c.httpServerHost
}

// HttpServerPort returns the built-in HTTP server port.
func (c *Config) HttpServerPort() int {
	return c.httpServerPort
}

// HttpServerMode returns the server mode.
func (c *Config) HttpServerMode() string {
	return c.httpServerMode
}

// HttpServerPassword returns the password for the user interface (optional).
func (c *Config) HttpServerPassword() string {
	return c.httpServerPassword
}

// OriginalsPath returns the originals.
func (c *Config) OriginalsPath() string {
	return c.originalsPath
}

// ImportPath returns the import directory.
func (c *Config) ImportPath() string {
	return c.importPath
}

// ExportPath returns the export directory.
func (c *Config) ExportPath() string {
	return c.exportPath
}

// DarktableCli returns the darktable-cli binary file name.
func (c *Config) DarktableCli() string {
	return c.darktableCli
}

// DatabaseDriver returns the database driver name.
func (c *Config) DatabaseDriver() string {
	return c.databaseDriver
}

// DatabaseDsn returns the database data source name (DSN).
func (c *Config) DatabaseDsn() string {
	return c.databaseDsn
}

// CachePath returns the path to the cache.
func (c *Config) CachePath() string {
	return c.cachePath
}

// ThumbnailsPath returns the path to the cached thumbnails.
func (c *Config) ThumbnailsPath() string {
	return c.CachePath() + "/thumbnails"
}

// AssetsPath returns the path to the assets.
func (c *Config) AssetsPath() string {
	return c.assetsPath
}

// TensorFlowModelPath returns the tensorflow model path.
func (c *Config) TensorFlowModelPath() string {
	return c.AssetsPath() + "/tensorflow"
}

// ServerPath returns the server assets path (public files, favicons, templates,...).
func (c *Config) ServerPath() string {
	return c.AssetsPath() + "/server"
}

// HttpTemplatesPath returns the server templates path.
func (c *Config) HttpTemplatesPath() string {
	return c.ServerPath() + "/templates"
}

// HttpFaviconsPath returns the favicons path.
func (c *Config) HttpFaviconsPath() string {
	return c.HttpPublicPath() + "/favicons"
}

// HttpPublicPath returns the public server path (//server/assets/*).
func (c *Config) HttpPublicPath() string {
	return c.ServerPath() + "/public"
}

// HttpPublicBuildPath returns the public build path (//server/assets/build/*).
func (c *Config) HttpPublicBuildPath() string {
	return c.HttpPublicPath() + "/build"
}

// Db returns the db connection.
func (c *Config) Db() *gorm.DB {
	if c.db == nil {
		c.connectToDatabase()
	}

	return c.db
}

// MigrateDb will start a migration process.
func (c *Config) MigrateDb() {
	db := c.Db()

	db.AutoMigrate(&models.File{},
		&models.Photo{},
		&models.Tag{},
		&models.Album{},
		&models.Location{},
		&models.Camera{},
		&models.Lens{},
		&models.Country{})
}

// ClientConfig returns a loaded and set configuration entity.
func (c *Config) ClientConfig() frontend.Config {
	db := c.Db()

	var cameras []*models.Camera

	type country struct {
		LocCountry     string
		LocCountryCode string
	}

	var countries []country

	db.Model(&models.Location{}).Select("DISTINCT loc_country_code, loc_country").Scan(&countries)

	db.Where("deleted_at IS NULL").Limit(1000).Order("camera_model").Find(&cameras)

	jsHash := fsutil.Hash(c.HttpPublicBuildPath() + "/app.js")
	cssHash := fsutil.Hash(c.HttpPublicBuildPath() + "/app.css")

	result := frontend.Config{
		"appName":    c.AppName(),
		"appVersion": c.AppVersion(),
		"debug":      c.Debug(),
		"cameras":    cameras,
		"countries":  countries,
		"jsHash":     jsHash,
		"cssHash":    cssHash,
	}

	return result
}
