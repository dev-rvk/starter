env "local" {
  src = "file://db/schema.sql"
  url = getenv("DATABASE_URL")
  dev = "docker://postgres/16-alpine/dev?search_path=public"
  migration {
    dir = "file://db/migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
env "production" {
  url = getenv("DATABASE_URL")
  migration {
    dir = "file://db/migrations"
  }
}
