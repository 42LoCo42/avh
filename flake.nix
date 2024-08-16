{
  description = "AvH video storage";

  outputs = { flake-utils, nixpkgs, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; }; in
      rec {
        packages.default = pkgs.buildGoModule rec {
          name = "avh";
          src = ./.;
          vendorHash = "sha256-TjzOEqbBP4qOCcZLtY0GPz0o5UhiSffiUTBwNVaLRXs=";

          CGO_ENABLED = "0";
          stripAllList = [ "bin" ];
          meta.mainProgram = name;
        };

        devShells.default = pkgs.mkShell {
          inputsFrom = [ packages.default ];
          packages = with pkgs; [
            air
            sqlite-interactive
          ];
        };
      });
}
