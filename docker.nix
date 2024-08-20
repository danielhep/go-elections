{ self ? (builtins.getFlake (toString ./.)) }:

let
  system = "x86_64-linux"; # Adjust this if you're targeting a different system
  pkgs = self.inputs.nixpkgs.legacyPackages.${system};
  devShell = self.devShells.${system}.default;

  app = pkgs.buildGoModule {
    pname = "elections";
    version = "0.1.0";
    src = ./.;
    vendorHash = null;
    # Use the Go version from the devShell
    buildInputs = [ devShell.config.languages.go.package ];
  };
in
pkgs.dockerTools.buildImage {
  name = "elections-app";
  tag = "latest";

  copyToRoot = pkgs.buildEnv {
    name = "image-root";
    paths = [
      app
      pkgs.cacert # Necessary for HTTPS requests
    ] ++ devShell.config.packages;
    pathsToLink = [ "/bin" "/etc" ];
  };

  config = {
    Cmd = [ "${app}/bin/elections" ];
    ExposedPorts = {
      "8080/tcp" = {};
    };
  };
}