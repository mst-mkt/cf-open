{
  description = "cf-open - open Cloudflare dashboard for your project from CLI";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
      version = self.shortRev or self.dirtyShortRev or "dev";
    in
    {
      packages = forAllSystems (pkgs: rec {
        default = cf-open;
        cf-open = pkgs.buildGoModule {
          pname = "cf-open";
          inherit version;
          src = ./.;
          vendorHash = "sha256-RrCXSBm3cEvtIvTHumJdLIwHe/lcYyP05UWUV3PkLIY=";
          subPackages = [ "cmd/cf-open" ];
          ldflags = [
            "-s"
            "-w"
            "-X main.version=${version}"
          ];
          env.CGO_ENABLED = 0;
          meta = {
            description = "Open Cloudflare dashboard for your project from CLI";
            homepage = "https://github.com/mst-mkt/cf-open";
            license = nixpkgs.lib.licenses.mit;
            mainProgram = "cf-open";
          };
        };
      });

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go
            golangci-lint
            just
            nodejs-slim_24
          ];
        };
      });
    };
}
