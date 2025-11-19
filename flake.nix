# Usage:
#   nix profile add github:asciimoth/connproxy
#   nix profile remove connproxy
#   nix shell github:asciimoth/connproxy
# Update: nix flake update
{
  description = "A simple net.Listener <-> net.Conn proxy";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    pre-commit-hooks.url = "github:cachix/pre-commit-hooks.nix";
    gomod2nix = {
      url = "github:tweag/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
    version = {
      url = "github:asciimoth/version";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };
  outputs = {
    self,
    nixpkgs,
    flake-utils,
    pre-commit-hooks,
    gomod2nix,
    version,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [ gomod2nix.overlays.default ];
      };



      release = pkgs.writeShellScriptBin "release" (builtins.readFile ./ci/release);

      tests = pkgs.writeShellScriptBin "tests" ''
        go test ./... "$@"
      '';

      checks = {
        pre-commit-check = pre-commit-hooks.lib.${system}.run {
          src = ./.;
          hooks = {
            commitizen.enable = true;
            typos.enable = true;
            typos-commit = {
              enable = true;
              description = "Find typos in commit message";
              entry = let script = pkgs.writeShellScript "typos-commit" ''
                typos "$1"
              ''; in builtins.toString script;
              stages = [ "commit-msg" ];
            };
            govet.enable = true;
            gofmt.enable = true;
            golangci-lint.enable = true;
            gotidy = {
              enable = true;
              description = "Makes sure go.mod matches the source code";
              entry = let script = pkgs.writeShellScript "gotidyhook" ''
                go mod tidy -v
                if [ -f "go.mod" ]; then
                  git add go.mod
                fi
                if [ -f "go.sum" ]; then
                  git add go.sum
                fi
              ''; in builtins.toString script;
              stages = [ "pre-commit" ];
            };
            version-check = {
              enable = true;
              description = "Makes sure SemVer values from all sources are matching";
              entry = let script = pkgs.writeShellScript "versionhook" ''
                version get
              ''; in builtins.toString script;
              stages = [ "pre-commit" ];
            };
          };
        };
      };
    in {
      devShells.default = pkgs.mkShell {
        inherit (checks.pre-commit-check) shellHook;
        buildInputs = with pkgs; [
          go
          golangci-lint
          commitizen
          goreleaser
          git-cliff
          govulncheck

          gomod2nix.packages.${system}.default
          version.packages.${system}.default

          ko

          typos

          yq
          jq

          release

          tests
        ];
      };


    });
}
