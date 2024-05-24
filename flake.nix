{
  description = "AvH video storage";

  inputs.obscura.url = "github:42loco42/obscura";

  outputs = { flake-utils, nixpkgs, obscura, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; }; in
      rec {
        packages.default = pkgs.buildGoModule {
          name = "avh";
          src = ./.;
          vendorHash = "sha256-C5YijILF5XFUA71O5KPf8RVro5njj/YMcHc5FFgdFpo=";

          nativeBuildInputs = [ obscura.packages.${pkgs.system}.jade ];
          preBuild = ''
            jade --writer -d jade .
          '';
        };

        devShells.default = pkgs.mkShell {
          inputsFrom = [ packages.default ];
          packages = with pkgs; [ gopls ];
        };

        packages.foo = pkgs.dockerTools.buildImage {
          name = "avh";
          tag = "latest";
          config.Cmd = [ (pkgs.lib.getExe packages.default) ];
        };
      });
}
