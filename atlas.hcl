env "development" {
  src = "file://db/schema.sql"
  dev = "sqlite://dev?mode=memory"
  url = "sqlite://flizix_dev.db"
}

env "production" {
  src = "file://db/schema.sql"
  dev = "sqlite://dev?mode=memory"
  url = "sqlite://flizix.db"
}