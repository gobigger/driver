package driver

import (
	_ "github.com/gobigger/driver/logger-default"
	_ "github.com/gobigger/driver/logger-file"
	_ "github.com/gobigger/driver/mutex-default"
	_ "github.com/gobigger/driver/mutex-memcache"
	_ "github.com/gobigger/driver/mutex-redis"
	_ "github.com/gobigger/driver/session-default"
	_ "github.com/gobigger/driver/session-redis"
	_ "github.com/gobigger/driver/session-memcache"
	_ "github.com/gobigger/driver/session-file"
	_ "github.com/gobigger/driver/session-memory"

	_ "github.com/gobigger/driver/cache-default"
	_ "github.com/gobigger/driver/cache-file"
	_ "github.com/gobigger/driver/cache-memory"
	_ "github.com/gobigger/driver/cache-memcache"
	_ "github.com/gobigger/driver/cache-redis"
	_ "github.com/gobigger/driver/data-postgres"
	_ "github.com/gobigger/driver/data-cockroach"
	_ "github.com/gobigger/driver/file-default"
	
	_ "github.com/gobigger/driver/plan-default"
	_ "github.com/gobigger/driver/event-default"
	_ "github.com/gobigger/driver/event-redis"
	_ "github.com/gobigger/driver/queue-default"
	_ "github.com/gobigger/driver/queue-redis"

	_ "github.com/gobigger/driver/http-default"
	_ "github.com/gobigger/driver/view-default"
	_ "github.com/gobigger/driver/socket-default"

)
