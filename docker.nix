{ pkgs }:

let
  system = "x86_64-linux"; # Adjust this if you're targeting a different system
  app = pkgs.buildGoModule {
    pname = "go-elections";
    version = "0.1.0";
    src = ./.;
    vendorHash = "sha256-RrAecmrGKxzJAE/r/Mt+nE+9ve9H9Msz3wWeAD3w1Lk=";
    # vendorHash = pkgs.lib.fakeHash;
    subPackages = [ "cmd/election-scraper" ];
  };
in
pkgs.dockerTools.buildImage {
  name = "go-elections";
  tag = "latest";

  copyToRoot = pkgs.buildEnv {
    name = "image-root";
    paths = [
      app
    ];
    pathsToLink = [ "/bin" "/etc" ];
  };

  config = {
    Cmd = [ "${app}/bin/go-elections" ];
    ExposedPorts = {
      "8080/tcp" = {};
    };
  };
}