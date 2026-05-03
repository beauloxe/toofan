{
  description = "A Nix flake for toofan";

  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = inputs @ {
    self,
    flake-parts,
    ...
  }:
    flake-parts.lib.mkFlake {inherit inputs;} {
      systems = ["x86_64-linux" "aarch64-linux"];
      perSystem = {
        config,
        self',
        inputs',
        pkgs,
        system,
        ...
      }: let
        toofan = let
          revision = self.shortRev or self.dirtyShortRev or "unknown";
        in
          pkgs.callPackage ./package.nix {version = revision;};
      in {
        packages = {
          inherit toofan;
          default = toofan;
        };
      };
    };
}
