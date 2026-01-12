data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./pkg/model",
    "--dialect", "postgres"
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url

  dev = "postgres://postgres:sebe1596@localhost:5432/go_task_api_db?sslmode=disable&search_path=public"

  migration {
    dir = "file://migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
