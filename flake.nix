{
  description = "GoUAE/golang.ae: The official site of the Go community in the UAE";

  outputs = inputs @ {flake-parts, ...}:
    flake-parts.lib.mkFlake {inherit inputs;} {
      imports = [{perSystem = {lib, ...}: {_module.args.l = lib // builtins;};}];
      systems = ["x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin"];
      perSystem = {
        l,
        pkgs,
        config,
        inputs',
        ...
      }: {
        devShells.default = pkgs.mkShell {
          packages = l.attrValues {
            inherit (pkgs) go gopls just;
          };
        };
      };
    };

  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };
}
