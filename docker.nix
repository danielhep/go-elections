{ pkgs }:

let
  system = "x86_64-linux"; # Adjust this if you're targeting a different system
  app = pkgs.buildGoModule {
    pname = "go-elections";
    version = "0.1.0";
    src = ./.;
    vendorHash = "sha256-mQXQ+T1kNDwzxamAcTSqbdJKTFIAXGvkZOvLJ0Y+z3E=";
    # vendorHash = pkgs.lib.fakeHash;
    subPackages = [ "cmd/election-scraper" "cmd/historical-import" "cmd/web" ];
  };
in
pkgs.dockerTools.buildImage {
  name = "go-elections";
  tag = "latest";

  copyToRoot = pkgs.buildEnv {
    name = "image-root";
    paths = [
      app
      pkgs.cacert  # Add SSL certificates
      pkgs.tzdata  # Add timezone data
    ];
    pathsToLink = [ "/bin" "/etc" ];
  };

  config = {
    Cmd = [ "election-scraper" ];
    ExposedPorts = {
      "8080/tcp" = {};
    };
  };
}
