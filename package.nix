{
  version,
  lib,
  buildGoModule,
}:
buildGoModule (finalAttrs: {
  pname = "toofan";
  inherit version;

  src = ./.;

  vendorHash = "sha256-YSjJ8NOL97hXZLnfGYIjoKmARv+gWOsv+5qkl9konnA=";

  ldflags = ["-s"];

  meta = {
    description = "A minimal, lightning-fast typing TUI for your terminal";
    homepage = "https://github.com/beauloxe/toofan";
    license = lib.licenses.mit;
    # maintainers = with lib.maintainers; [];
    mainProgram = "toofan";
  };
})
