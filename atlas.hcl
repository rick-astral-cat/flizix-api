env "local" {
  src = "file://db/schema.sql"
  dev = "sqlite://dev?mode=memory"
  url = "sqlite://flizix.db"
}
