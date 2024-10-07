# King County Elections Parser

King County and the State of Washington publish election data as CSVs on their websites. The goal of this project is to provide tools for parsing these files into a database for displaying election results. A server rendered web application is also provided to view the data. For other frontends, use a program such as [PostgREST](https://postgrest.org/) or [Postgraphile](https://postgraphile.com/).

### Web Application
The web applition is a simple frontend that connects to the database and displays election results. There are simple graphs displayed for each contest. 

### Scraper
The scraper is a program that connects to the King County and State of Washington websites and downloads the CSV files. It continusally pulls the CSV file and hashes it to check if it has changed. If it has changed, it parses the CSV and inserts the new vote tallies into the database.

### Importer
The importer is a command line tool that can be run on a directorry containing the CSV files downloaded from King County or State of Washington elections websites. It is able to prase the filenames to determine the dates and whether the file came from the state or county. You must specify some other parameters which can be seen in the help text.

## Development
The development environment is provided by [Nix](https://nixos.org/) using flakes and [devenv](https://devenv.sh/). The development environment is defined in `devenv.nix`.  Run `devenv shell` to enter the development environment. `devenv up` will start the Postgres server. 

Run each of the three applications by running `go run ./cmd/<app>`. For example, to run the web application, run `go run ./cmd/web`.

Additionally, the web application templates are written using [Templ](https://templ.dev/). To update the templates, run `templ generate -watch`.