{
  description = "AvH video storage";

  inputs.obscura.url = "github:42loco42/obscura";

  outputs = { flake-utils, nixpkgs, obscura, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; }; in
      rec {
        packages.default = pkgs.buildGoModule rec {
          name = "avh";
          src = ./.;
          vendorHash = "sha256-C5YijILF5XFUA71O5KPf8RVro5njj/YMcHc5FFgdFpo=";

          nativeBuildInputs = [ obscura.packages.${pkgs.system}.jade ];
          preBuild = ''
            jade --writer -d jade .
          '';

          meta.mainProgram = name;
        };

        devShells.default = pkgs.mkShell {
          inputsFrom = [ packages.default ];
          packages = with pkgs; [ gopls ];
        };
      });
}
