package config

import (
	"time"
)

var DEFAULT_CONFIG = Config{
	Server: ServerConfig{
		Port:    8000,
		Address: "127.0.0.1",
		Debug:   false,
	},
	Logger: LoggerConfig{
		Level:          "info",
		Target:         "stdout",
		Filename:       "",
		MaxFileSizeMB:  1,
		MaxFileBackups: 3,
		MaxFileAgeDays: 28,
		FileCompress:   true,
	},
	Watcher: WatcherConfig{
		Wait:        1 * time.Minute,
		JournalWait: 24 * time.Hour,
	},
	Database: DatabaseConfig{
		Host:     "127.0.0.1",
		Port:     5432,
		Username: "kitis",
		Password: "qwertyuiop123",
		Database: "kitisapi",
		Timeout:  5 * time.Second,
	},
	Parser: ParserConfig{
		Timeout:   5 * time.Second,
		BaseURI:   "http://94.72.18.202:8083",
		URIsuffix: ".htm",
		URIprefixes: map[string]string{
			"index":    "/index",
			"bells":    "/bells",
			"group":    "/cg",
			"lecturer": "/cp",
			"room":     "/ca",
		},
		Bells: BellsConfig{
			Monday: map[int]string{
				-1: "15:20-15:50",
				0:  "00:00-00:00",
				1:  "08:15-09:00",
				2:  "09:10-10:30",
				3:  "10:40-12:00",
				4:  "12:20-13:40",
				5:  "13:50-15:10",
				6:  "16:00-17:20",
				7:  "17:30-18:50",
				8:  "19:00-20:20",
				9:  "20:30-21:50",
			},
			General: map[int]string{
				0: "00:00-00:00",
				1: "08:30-10:00",
				2: "10:10-11:40",
				3: "12:10-13:40",
				4: "13:50-15:20",
				5: "15:30-17:00",
				6: "17:10-18:40",
				7: "18:50-20:20",
				8: "20:30-22:00",
				9: "22:10-23:40",
			},
		},
	},
}
