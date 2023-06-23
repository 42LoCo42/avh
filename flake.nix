{
  description = "AvH video storage";

  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        defaultPackage = pkgs.buildGoModule {
          pname = "avh";
          version = "1";
          src = ./.;

          vendorSha256 = "sha256-6J1qZ10MVSljIbmixZ0KD//Vr+txbZz4Ct5dqPEM76I=";
        };

        devShell = pkgs.mkShell {
          packages = with pkgs; [
            bashInteractive
            go
            gopls
          ];
        };
      });
}
