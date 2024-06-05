{
  description = "GoUAE/Hummus: A Whatsapp -> Discord Read-only Bridge for the GoUAE Community";

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
            inherit (pkgs) go gopls just imagemagick pkg-config;
          };
        };
      };
    };

  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };
}
