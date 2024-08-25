{
  description = "AvH video storage";

  outputs = { flake-utils, nixpkgs, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; }; in
      rec {
        packages.default = pkgs.buildGoModule rec {
          name = "avh";
          src = ./.;
          vendorHash = "sha256-2lzghQws/RghTv2EnhELLVr63yff+9sSmDlia2jH3QI=";

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
