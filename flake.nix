{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
    devenv.url = "github:cachix/devenv";
    devenv.inputs.nixpkgs.follows = "nixpkgs";
  };

  # nixConfig = {
  #   extra-trusted-public-keys = "devenv.cachix.org-1:w1cLUi8dv3hnoSPGAuibQv+f9TZLr6cv/Hm9XgU50cw=";
  #   extra-substituters = "https://devenv.cachix.org";
  # };

  outputs = { self, nixpkgs, devenv, systems, ... } @ inputs:
    let
      forEachSystem = nixpkgs.lib.genAttrs (import systems);
      overlay = final: prev: {
        go = prev.go_1_23;
        buildGoModule = prev.buildGoModule.override {
          go = final.go;
        };
      };
      pkgsForSystem = system: import nixpkgs {
        inherit system;
        overlays = [ overlay ];
      };
    in
    {
      overlays.default = overlay;
      packages = forEachSystem (system: 
      let
        pkgs = pkgsForSystem system;
      in {
        devenv-up = self.devShells.${system}.default.config.procfileScript;
        docker = pkgs.callPackage ./docker.nix { pkgs };
      });

      devShells = forEachSystem
        (system:
          let
            pkgs = pkgsForSystem system;
          in
          {
            default = devenv.lib.mkShell {
              inherit inputs pkgs;
              modules = [
                ./devenv.nix
              ];
            };
          });
    };
}
