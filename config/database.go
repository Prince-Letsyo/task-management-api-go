package config

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Prince-Letsyo/task-management-api-go/internal/domain"
	"github.com/oarkflow/log"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

type DatabaseDriver struct {
	Driver            string        `yaml:"driver" env:"DB_DRIVER"`
	Host              string        `json:"host" yaml:"host" env:"DB_HOST" env-default:"127.0.0.1"`
	Username          string        `yaml:"username" env:"DB_USER"`
	Password          string        `json:"password" yaml:"password" env:"DB_PASS" env-default:"postgres"`
	DBName            string        `yaml:"db_name" env:"DB_NAME"`
	Port              int           `json:"port" yaml:"port" env:"DB_PORT" env-default:"5432"`
	Connections       int           `yaml:"connections" env:"DB_CONNECTIONS"`
	SslMode           string        `json:"ssl_mode" yaml:"sslMode" env:"DB_SSL_MODE" env-default:"disable"`
	MaxConns          int           `json:"max_conns" yaml:"maxConns" env:"MAX_CONNS" env-default:"10"`
	MinConns          int           `json:"min_conns" yaml:"minConns" env:"MIN_CONNS" env-default:"2"`
	MaxConnLifetime   time.Duration `json:"max_conn_lifetime" yaml:"maxConnLifetime" env:"MAX_CONN_LIFETIME" env-default:"10m"`
	MaxConnIdleTime   time.Duration `json:"max_conn_idle_time" yaml:"maxConnIdleTime" env:"MAX_CONN_IDLE_TIME" env-default:"1m"`
	HealthCheckPeriod time.Duration `json:"health_check_period" yaml:"healthCheckPeriod" env:"HEALTH_CHECK_PERIOD" env-default:"10s"`
}

type DatabaseConfig struct {
	*gorm.DB
	Drivers map[string]DatabaseDriver `yaml:"drivers"`
	Default DatabaseDriver            `yaml:"default" env:"DEFAULT_DB_DRIVER"`
}

func (d *DatabaseConfig) Setup() error {
	//nolint:wsl,lll
	var err error //nolint:wsl
	connectionString := ""
	if d.DB != nil {
		return nil
	}
	gormLogger := New(&log.DefaultLogger, logger.Config{
		LogLevel: 0,
	}, false)
	newLogger := gormLogger.LogMode(logger.Info)
	switch d.Default.Driver {
	case "postgres":
		connectionString = fmt.Sprintf(
			"host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
			d.Default.Host,
			d.Default.Port,
			d.Default.Username,
			d.Default.DBName,
			d.Default.Password,
			d.Default.SslMode,
		)
		d.DB, err = gorm.Open(postgres.Open(connectionString), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
			Logger:                                   newLogger,
		})

	default:
		connectionString = fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
			d.Default.Username,
			d.Default.Password,
			d.Default.Host,
			d.Default.Port,
			d.Default.DBName,
		)
		d.DB, err = gorm.Open(mysql.Open(connectionString), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
			Logger:                                   newLogger,
		})
	}
	if err != nil {
		panic(err)
	}
	if err := d.DB.Use(
		dbresolver.Register(dbresolver.Config{}).
			SetConnMaxLifetime(d.Default.MaxConnLifetime).
			SetMaxIdleConns(100).
			SetMaxOpenConns(100).
			SetConnMaxIdleTime(d.Default.MaxConnIdleTime),
	); err != nil {
		return err
	}
	return nil //nolint:wsl
}

func (d *DatabaseConfig) Paginate(filters *domain.Filters) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if filters.Offset <= 0 {
			filters.Offset = 1
		}
		switch {
		case filters.Limit > domain.MaxPageLimit:
			filters.Limit = domain.MaxPageLimit
		case filters.Limit <= 0:
			filters.Limit = domain.DefaultPageLimit
		}

		dbClone := db.Session(&gorm.Session{})

		var total int64
		dbClone.Count(&total)

		filters.Total = int64(math.Ceil(float64(total) / float64(filters.Limit)))

		filters.Offset = (filters.Offset - 1) * filters.Limit
		return db.Offset(int(filters.Offset)).Limit(int(filters.Limit))
	}
}

func New(logger *log.Logger, config logger.Config, slient bool) logger.Interface {
	return &gormLogger{
		Log:    logger,
		Config: config,
		Slient: slient,
	}
}

type gormLogger struct {
	Log    *log.Logger
	Config logger.Config
	Slient bool
}

func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := gormLogger{Log: l.Log}
	switch level {
	case logger.Silent:
		newLogger.Slient = true
	case logger.Error:
		newLogger.Log.SetLevel(log.ErrorLevel)
	case logger.Warn:
		newLogger.Log.SetLevel(log.WarnLevel)
	case logger.Info:
		newLogger.Log.SetLevel(log.InfoLevel)
	}

	return &newLogger
}

func (l *gormLogger) Info(ctx context.Context, format string, args ...interface{}) {
	l.Log.Info().Msgf(format, args...)
}

func (l *gormLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	l.Log.Warn().Msgf(format, args...)
}

func (l *gormLogger) Error(ctx context.Context, format string, args ...interface{}) {
	l.Log.Error().Msgf(format, args...)
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.Slient {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.Log.Level >= log.ErrorLevel:
		sql, rows := fc()
		if rows == -1 {
			l.Log.Error().Caller(1).Err(err).Dur("elapsed", elapsed).Str("sql", sql).Msg("")
		} else {
			l.Log.Error().Caller(1).Err(err).Dur("elapsed", elapsed).Str("sql", sql).Int64("rows", rows).Msg("")
		}
	case elapsed > l.Config.SlowThreshold && l.Config.SlowThreshold != 0 && l.Log.Level >= log.WarnLevel:
		sql, rows := fc()
		if rows == -1 {
			l.Log.Warn().Caller(1).Err(err).Dur("elapsed", elapsed).Str("sql", sql).Msgf("SLOW SQL >= %v", l.Config.SlowThreshold)
		} else {
			l.Log.Warn().Caller(1).Err(err).Dur("elapsed", elapsed).Str("sql", sql).Int64("rows", rows).Msgf("SLOW SQL >= %v", l.Config.SlowThreshold)
		}
	case l.Log.Level == log.InfoLevel:
		sql, rows := fc()
		if rows == -1 {
			l.Log.Info().Caller(1).Err(err).Dur("elapsed", elapsed).Str("sql", sql).Msg("")
		} else {
			l.Log.Info().Caller(1).Err(err).Dur("elapsed", elapsed).Str("sql", sql).Int64("rows", rows).Msg("")
		}
	}
}
