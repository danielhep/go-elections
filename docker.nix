{ self ? (builtins.getFlake (toString ./.)) }:

let
  system = "x86_64-linux"; # Adjust this if you're targeting a different system
  pkgs = self.inputs.nixpkgs.legacyPackages.${system};

  app = pkgs.buildGoModule {
    pname = "elections";
    version = "0.1.0";
    src = ./.;
    vendorHash = null;
    buildInputs = [ pkgs.go_1_23 ];
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
      gopls
      gotools
      go-tools
    ];
    pathsToLink = [ "/bin" "/etc" ];
  };

  config = {
    Cmd = [ "${app}/bin/elections" ];
    ExposedPorts = {
      "8080/tcp" = {};
    };
  };
}