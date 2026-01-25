package config

import (
	"context"
	"fmt"
	"time"

	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/oarkflow/log"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseDriver struct {
	Driver            string        `yaml:"driver" env:"DB_DRIVER"`
	Host              string        `yaml:"host" env:"DB_HOST" env-default:"127.0.0.1"`
	Username          string        `yaml:"username" env:"DB_USER"`
	Password          string        `yaml:"password" env:"DB_PASS"`
	DBName            string        `yaml:"db_name" env:"DB_NAME"`
	Port              int           `yaml:"port" env:"DB_PORT" env-default:"5432"`
	SslMode           string        `yaml:"ssl_mode" env:"DB_SSL_MODE" env-default:"disable"`
	MaxConns          int           `yaml:"max_conns" env:"MAX_CONNS" env-default:"25"`
	MinConns          int           `yaml:"min_conns" env:"MIN_CONNS" env-default:"5"`
	MaxConnLifetime   time.Duration `yaml:"max_conn_lifetime" env:"MAX_CONN_LIFETIME" env-default:"10m"`
	MaxConnIdleTime   time.Duration `yaml:"max_conn_idle_time" env:"MAX_CONN_IDLE_TIME" env-default:"10m"`
	HealthCheckPeriod time.Duration `yaml:"health_check_period" env:"HEALTH_CHECK_PERIOD" env-default:"1m"`
}

type DatabaseConfig struct {
	*gorm.DB
	Drivers map[string]DatabaseDriver `yaml:"drivers"`
	Default DatabaseDriver            `yaml:"default"`
}

// Setup initializes the database connection
func (d *DatabaseConfig) Setup() error {
	if d.DB != nil {
		return nil // Already initialized
	}

	var dialector gorm.Dialector

	switch d.Default.Driver {
	case "postgres", "postgresql":
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
			d.Default.Host,
			d.Default.Port,
			d.Default.Username,
			d.Default.DBName,
			d.Default.Password,
			d.Default.SslMode,
		)
		dialector = postgres.Open(dsn)

	case "mysql":
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			d.Default.Username,
			d.Default.Password,
			d.Default.Host,
			d.Default.Port,
			d.Default.DBName,
		)
		dialector = mysql.Open(dsn)

	default:
		return fmt.Errorf("unsupported database driver: %s", d.Default.Driver)
	}

	gormLogger := NewGORMLogger(&log.DefaultLogger, logger.Config{
		SlowThreshold:             time.Second,
		LogLevel:                  logger.Info,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	})

	db, err := gorm.Open(dialector, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   gormLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(d.Default.MaxConns)
	sqlDB.SetMaxIdleConns(d.Default.MinConns)
	sqlDB.SetConnMaxLifetime(d.Default.MaxConnLifetime)
	sqlDB.SetConnMaxIdleTime(d.Default.MaxConnIdleTime)

	// Background health check
	if d.Default.HealthCheckPeriod > 0 {
		go func() {
			ticker := time.NewTicker(d.Default.HealthCheckPeriod)
			defer ticker.Stop()
			for range ticker.C {
				if err := sqlDB.Ping(); err != nil {
					log.DefaultLogger.Error().
						Err(err).
						Msg("Database health check failed")
				}
			}
		}()
	}

	d.DB = db
	return nil
}

// Close gracefully shuts down the connection pool
func (d *DatabaseConfig) Close() error {
	if d.DB == nil {
		return nil
	}
	sqlDB, _ := d.DB.DB() // ignore error — we're closing anyway
	return sqlDB.Close()
}

// Paginate returns a GORM scope for pagination
func (d *DatabaseConfig) Paginate(filters *pkg.Filters) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Sanitize inputs
		if filters.Limit <= 0 {
			filters.Limit = pkg.DefaultPageLimit
		}
		if filters.Limit > pkg.MaxPageLimit {
			filters.Limit = pkg.MaxPageLimit
		}
		if filters.Offset < 1 {
			filters.Offset = 1
		}

		// Calculate offset (page-based → zero-based)
		offset := (filters.Offset - 1) * filters.Limit

		// Count total records for pagination metadata
		var total int64
		db.Count(&total)
		filters.Total = total // Total records, not pages!

		return db.Offset(int(offset)).Limit(int(filters.Limit))
	}
}

// === Custom GORM Logger ===

type gormLogger struct {
	logger *log.Logger
	config logger.Config
}

// NewGORMLogger creates a bridge from oarkflow/log to GORM's logger interface
func NewGORMLogger(l *log.Logger, config logger.Config) logger.Interface {
	return &gormLogger{
		logger: l,
		config: config,
	}
}

func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newConfig := l.config
	newConfig.LogLevel = level
	return &gormLogger{
		logger: l.logger,
		config: newConfig,
	}
}

func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel >= logger.Info {
		l.logger.Info().Msgf(msg, data...)
	}
}

func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel >= logger.Warn {
		l.logger.Warn().Msgf(msg, data...)
	}
}

func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel >= logger.Error {
		l.logger.Error().Msgf(msg, data...)
	}
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.config.LogLevel == logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.config.LogLevel >= logger.Error:
		l.logger.Error().
			Err(err).
			Dur("elapsed", elapsed).
			Str("sql", sql).
			Int64("rows", rows).
			Msg("GORM query error")

	case elapsed > l.config.SlowThreshold && l.config.SlowThreshold != 0 && l.config.LogLevel >= logger.Warn:
		l.logger.Warn().
			Dur("elapsed", elapsed).
			Str("sql", sql).
			Int64("rows", rows).
			Msgf("SLOW SQL > %v", l.config.SlowThreshold)

	case l.config.LogLevel >= logger.Info:
		l.logger.Info().
			Dur("elapsed", elapsed).
			Str("sql", sql).
			Int64("rows", rows).
			Msg("GORM query executed")
	}
}
