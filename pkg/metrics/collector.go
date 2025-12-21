package metrics

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// Collector periodically collects business metrics from the database
type Collector struct {
	db     *sqlx.DB
	logger *logrus.Logger
	stopCh chan struct{}
}

// NewCollector creates a new metrics collector
func NewCollector(db *sqlx.DB, logger *logrus.Logger) *Collector {
	return &Collector{
		db:     db,
		logger: logger,
		stopCh: make(chan struct{}),
	}
}

// Start begins collecting metrics periodically
func (c *Collector) Start(interval time.Duration) {
	// Collect immediately on start
	c.collect()

	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.collect()
			case <-c.stopCh:
				ticker.Stop()
				return
			}
		}
	}()

	c.logger.WithField("interval", interval).Info("Metrics collector started")
}

// Stop stops the metrics collector
func (c *Collector) Stop() {
	close(c.stopCh)
	c.logger.Info("Metrics collector stopped")
}

// collect gathers all business metrics from the database
func (c *Collector) collect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Products count
	var productsCount int64
	if err := c.db.GetContext(ctx, &productsCount, "SELECT COUNT(*) FROM products"); err != nil {
		c.logger.WithError(err).Warn("Failed to collect products count")
	} else {
		ProductsTotal.Set(float64(productsCount))
	}

	// Offers count
	var offersCount int64
	if err := c.db.GetContext(ctx, &offersCount, "SELECT COUNT(*) FROM offers WHERE is_available = true"); err != nil {
		c.logger.WithError(err).Warn("Failed to collect offers count")
	} else {
		OffersTotal.Set(float64(offersCount))
	}

	// Active users (logged in within last 24h) - approximate via orders
	var activeUsers int64
	if err := c.db.GetContext(ctx, &activeUsers, `
		SELECT COUNT(DISTINCT user_id) FROM orders 
		WHERE created_at > NOW() - INTERVAL '24 hours'
	`); err != nil {
		c.logger.WithError(err).Warn("Failed to collect active users count")
	} else {
		UsersActive.Set(float64(activeUsers))
	}

	// DB connections
	stats := c.db.Stats()
	DBConnectionsActive.Set(float64(stats.InUse))

	c.logger.Debug("Metrics collected successfully")
}
