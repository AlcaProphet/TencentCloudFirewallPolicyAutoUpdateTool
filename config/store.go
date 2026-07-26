package config

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
)

// Store SQLite 配置持久化
type Store struct {
	db *sql.DB
}

// SyncLog 同步日志记录
type SyncLog struct {
	Timestamp time.Time `json:"timestamp"`
	Target    string    `json:"target"`
	Domain    string    `json:"domain"`
	Result    string    `json:"result"` // success / failed / skipped
	Added     int       `json:"added"`
	Deleted   int       `json:"deleted"`
	Error     string    `json:"error"`
}

// OpenStore 打开或创建 SQLite 数据库
func OpenStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// WAL 模式 + busy_timeout
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("设置 WAL 模式失败: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("设置 busy_timeout 失败: %w", err)
	}

	s := &Store{db: db}
	if err := s.initTables(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Close 关闭数据库
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) initTables() error {
	schema := `
CREATE TABLE IF NOT EXISTS targets (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	cloud_type TEXT NOT NULL,
	region TEXT NOT NULL,
	resource_id TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS rules (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	host TEXT NOT NULL,
	protocol TEXT NOT NULL,
	ports TEXT NOT NULL,
	action TEXT NOT NULL DEFAULT 'ACCEPT',
	targets TEXT DEFAULT '',
	comment TEXT DEFAULT ''
);
CREATE TABLE IF NOT EXISTS settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS sync_logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	target TEXT,
	domain TEXT,
	result TEXT,
	added INTEGER DEFAULT 0,
	deleted INTEGER DEFAULT 0,
	error TEXT DEFAULT ''
);
`
	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("初始化表结构失败: %w", err)
	}
	return nil
}

// GetSettings 获取全局设置
func (s *Store) GetSettings() (map[string]string, error) {
	rows, err := s.db.Query("SELECT key, value FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		settings[k] = v
	}
	return settings, rows.Err()
}

// SetSetting 设置单项配置
func (s *Store) SetSetting(key, value string) error {
	_, err := s.db.Exec(
		"INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)",
		key, value,
	)
	return err
}

// GetTargets 获取所有目标
func (s *Store) GetTargets() ([]TargetConfig, error) {
	rows, err := s.db.Query("SELECT id, cloud_type, region, resource_id FROM targets ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []TargetConfig
	for rows.Next() {
		var t TargetConfig
		var ct string
		if err := rows.Scan(&t.ID, &ct, &t.Region, &t.ResourceID); err != nil {
			return nil, err
		}
		t.CloudType = CloudType(ct)
		targets = append(targets, t)
	}
	return targets, rows.Err()
}

// AddTarget 添加目标
func (s *Store) AddTarget(t TargetConfig) error {
	_, err := s.db.Exec(
		"INSERT INTO targets (cloud_type, region, resource_id) VALUES (?, ?, ?)",
		string(t.CloudType), t.Region, t.ResourceID,
	)
	return err
}

// DeleteTarget 删除目标
func (s *Store) DeleteTarget(id int) error {
	_, err := s.db.Exec("DELETE FROM targets WHERE id = ?", id)
	return err
}

// GetRules 获取所有域名规则
func (s *Store) GetRules() ([]DomainRule, error) {
	rows, err := s.db.Query("SELECT id, host, protocol, ports, action, targets, comment FROM rules ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []DomainRule
	for rows.Next() {
		var r DomainRule
		var targets string
		if err := rows.Scan(&r.ID, &r.Host, &r.Protocol, &r.Ports, &r.Action, &targets, &r.Comment); err != nil {
			return nil, err
		}
		if targets != "" {
			var nums []int
			if err := json.Unmarshal([]byte(targets), &nums); err == nil {
				r.Targets = nums
			}
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

// AddRule 添加域名规则
func (s *Store) AddRule(r DomainRule) error {
	targetsJSON, _ := json.Marshal(r.Targets)
	_, err := s.db.Exec(
		"INSERT INTO rules (host, protocol, ports, action, targets, comment) VALUES (?, ?, ?, ?, ?, ?)",
		r.Host, r.Protocol, r.Ports, r.Action, string(targetsJSON), r.Comment,
	)
	return err
}

// DeleteRule 删除域名规则
func (s *Store) DeleteRule(id int) error {
	_, err := s.db.Exec("DELETE FROM rules WHERE id = ?", id)
	return err
}

// UpdateTarget 更新目标
func (s *Store) UpdateTarget(id int, t TargetConfig) error {
	_, err := s.db.Exec(
		"UPDATE targets SET cloud_type = ?, region = ?, resource_id = ? WHERE id = ?",
		string(t.CloudType), t.Region, t.ResourceID, id,
	)
	return err
}

// UpdateRule 更新域名规则
func (s *Store) UpdateRule(id int, r DomainRule) error {
	targetsJSON, _ := json.Marshal(r.Targets)
	_, err := s.db.Exec(
		"UPDATE rules SET host = ?, protocol = ?, ports = ?, action = ?, targets = ?, comment = ? WHERE id = ?",
		r.Host, r.Protocol, r.Ports, r.Action, string(targetsJSON), r.Comment, id,
	)
	return err
}

// ClearAll 清空所有配置（用于配置导入前重置）
func (s *Store) ClearAll() error {
	_, err := s.db.Exec("DELETE FROM targets; DELETE FROM rules; DELETE FROM settings;")
	return err
}

// WithTransaction 在事务中执行操作，失败自动回滚
func (s *Store) WithTransaction(fn func(tx *sql.Tx) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// BatchAddTargets 批量添加目标
func (s *Store) BatchAddTargets(targets []TargetConfig) error {
	for _, t := range targets {
		if err := s.AddTarget(t); err != nil {
			return err
		}
	}
	return nil
}

// BatchAddRules 批量添加规则
func (s *Store) BatchAddRules(rules []DomainRule) error {
	for _, r := range rules {
		if err := s.AddRule(r); err != nil {
			return err
		}
	}
	return nil
}

// AddSyncLog 添加同步日志
func (s *Store) AddSyncLog(log SyncLog) error {
	_, err := s.db.Exec(
		"INSERT INTO sync_logs (timestamp, target, domain, result, added, deleted, error) VALUES (?, ?, ?, ?, ?, ?, ?)",
		log.Timestamp, log.Target, log.Domain, log.Result, log.Added, log.Deleted, log.Error,
	)
	if err != nil {
		return err
	}
	// 保留最近 1000 条
	_, err = s.db.Exec("DELETE FROM sync_logs WHERE id NOT IN (SELECT id FROM sync_logs ORDER BY id DESC LIMIT 1000)")
	return err
}

// GetSyncLogs 获取最近 N 条同步日志
func (s *Store) GetSyncLogs(limit int) ([]SyncLog, error) {
	rows, err := s.db.Query(
		"SELECT timestamp, target, domain, result, added, deleted, error FROM sync_logs ORDER BY id DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []SyncLog
	for rows.Next() {
		var l SyncLog
		if err := rows.Scan(&l.Timestamp, &l.Target, &l.Domain, &l.Result, &l.Added, &l.Deleted, &l.Error); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

// LoadConfig 从 SQLite 构建 Config
func (s *Store) LoadConfig() (*Config, error) {
	targets, err := s.GetTargets()
	if err != nil {
		return nil, err
	}
	rules, err := s.GetRules()
	if err != nil {
		return nil, err
	}
	settings, err := s.GetSettings()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Targets:          targets,
		DomainRules:      rules,
		Tag:              "auto-dns",
		Interval:         5 * time.Minute,
		DNS:              "8.8.8.8:53",
		DNSTimeout:       10 * time.Second,
		DNSFailThreshold: 5,
		LogLevel:         "info",
		WebUIPort:        9090,
		Mode:             "webui",
		TCAccessID:       settings["tc_access_id"],
		TCAccessKey:      settings["tc_access_key"],
		AliAccessID:      settings["ali_access_id"],
		AliAccessKey:     settings["ali_access_key"],
	}

	if v := settings["tag"]; v != "" {
		cfg.Tag = v
	}
	if v := settings["interval"]; v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Interval = d
		}
	}
	if v := settings["dns"]; v != "" {
		cfg.DNS = v
	}
	if v := settings["log_level"]; v != "" {
		cfg.LogLevel = v
	}
	if v := settings["webui_port"]; v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.WebUIPort = n
		}
	}
	if v := settings["dns_fail_threshold"]; v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.DNSFailThreshold = n
		}
	}

	return cfg, nil
}
