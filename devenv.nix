{ pkgs, ... }:

{
  # Packages to install
  packages = with pkgs; [
    go_1_23
    gopls
    gotools
    go-tools
    delve
  ] ++ lib.optionals pkgs.stdenv.isDarwin [
    pkgs.darwin.apple_sdk.frameworks.CoreFoundation
    pkgs.darwin.apple_sdk.frameworks.Security
    pkgs.darwin.apple_sdk.frameworks.SystemConfiguration
    pkgs.darwin.apple_sdk.frameworks.Cocoa
  ];

  # Enable Go language support
  languages.go.enable = true;

  # Set up environment variables
  env = {
    PG_URL = "postgres://postgres@localhost:5432/elections?sslmode=disable";
    STATE_DATA = "https://results.vote.wa.gov/results/20240806/export/20240806_AllState.csv";
    COUNTY_DATA = "https://aqua.kingcounty.gov/elections/2024/aug-primary/webresults.csv";
    GOATCOUNTER_URL = "https://danielhep.goatcounter.com/count";
  };

  # Shell configuration
  enterShell = ''
    echo "Go development environment loaded!"
    echo "Go version: $(go version)"
    echo "Postgres URL: $PG_URL"
  '';

  scripts.run.exec = ''
    go run src/*.go
  '';

  scripts.drop-all.exec = ''
    psql -U postgres -d elections <<EOF
    DO \$\$
    DECLARE
      r RECORD;
    BEGIN
      FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
        EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
      END LOOP;
    END
    \$\$;
    EOF
    echo "All tables have been dropped."
  '';

  # Project-specific configurations
  dotenv.enable = true;
  dotenv.filename = ".env";

  # Pre-commit hooks
  pre-commit.hooks = {
    gofmt.enable = true;
    golangci-lint.enable = true;
  };

  # Redis service configuration
  services.postgres = {
    enable = true;
    listen_addresses = "127.0.0.1";
    initialScript = ''
      CREATE USER postgres SUPERUSER;
      CREATE DATABASE elections WITH OWNER postgres;
    '';
  };

  # Add any other project-specific configurations here
}