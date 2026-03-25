{
  description = "Trust Wallet Core - C++ FFI with Rust core";

  inputs = {
    nixpkgs.url     = "github:nixos/nixpkgs/nixos-unstable";
    nixpkgs-protobuf.url = "github:nixos/nixpkgs/nixos-23.11";
    flake-parts.url = "github:hercules-ci/flake-parts";
    nci.url         = "github:yusdacra/nix-cargo-integration";
  };

  outputs = inputs:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.nci.flakeModule ];

      systems = [ "x86_64-linux" "aarch64-linux" ];

      perSystem = { config, pkgs, system, ... }:
        let
          inherit (pkgs) lib;

          stdenv = pkgs.llvmPackages.stdenv;

          # Use protobuf 3.20 from older nixpkgs for compatibility with the C++ plugins
          # Newer protobuf (22+) uses Abseil and returns absl::string_view instead of std::string
          protobuf-pkg = inputs.nixpkgs-protobuf.legacyPackages.${system}.protobuf3_20;

          # Custom C++ protobuf codegen plugins
          protobuf-plugins = stdenv.mkDerivation {
            pname    = "wallet-core-protobuf-plugins";
            version  = "0.1.0";
            src      = ./protobuf-plugin;

            nativeBuildInputs = [ pkgs.cmake protobuf-pkg ];

            # protobuf 3.20 doesn't ship cmake config files, use module mode
            postPatch = ''
              substituteInPlace CMakeLists.txt \
                --replace-fail 'find_package(Protobuf CONFIG REQUIRED PATH ''${PREFIX}/lib/pkgconfig)' \
                               'find_package(Protobuf REQUIRED)'
            '';

            cmakeFlags = [
              "-DProtobuf_PROTOC_EXECUTABLE=${protobuf-pkg}/bin/protoc"
            ];

            meta = {
              description = "Wallet Core protobuf compiler plugins";
              license     = lib.licenses.asl20;
            };
          };

          codegen-tools = pkgs.runCommand "wallet-core-codegen" {
            nativeBuildInputs = [ pkgs.makeWrapper ];
          } ''
            mkdir -p $out/bin
            for bin in ${./codegen}/bin/*; do
              install -m755 "$bin" "$out/bin/$(basename "$bin")"
              wrapProgram "$out/bin/$(basename "$bin")" \
                --prefix PATH : ${lib.makeBinPath [ pkgs.ruby pkgs.which ]}
            done
          '';

          # Rust build needs: rust/ workspace + src/proto/ (for tw_proto)
          rustSourceFilter = path: type:
            let
              rel = lib.removePrefix (toString ./.) (toString path);
            in
              lib.hasPrefix "/rust/"     rel ||
              lib.hasPrefix "/src/proto" rel ||
              rel == "/rust" || rel == "/src" || rel == "";

          # C++ build needs headers, sources, cmake, and generated proto files
          cppSourceFilter = path: type:
            let
              rel      = lib.removePrefix (toString ./.) (toString path);
              extMatch = ext: lib.hasSuffix ext (baseNameOf path);
            in
              type == "directory" ||
              extMatch ".cpp" || extMatch ".cc"  || extMatch ".c"    ||
              extMatch ".h"   || extMatch ".hpp" || extMatch ".proto"  ||
              extMatch ".cmake" || extMatch ".txt" || extMatch ".json"  ||
              lib.hasPrefix "/src/"          rel ||
              lib.hasPrefix "/include/"      rel ||
              lib.hasPrefix "/trezor-crypto/" rel ||
              lib.hasPrefix "/cmake/"        rel ||
              lib.hasPrefix "/jni/cpp/"      rel;

          mkWalletCoreFFI = rustLib:
            stdenv.mkDerivation {
              pname   = "wallet-core-ffi";
              version = "0.1.0";

              src = lib.cleanSourceWith {
                src    = ./.;
                filter = cppSourceFilter;
              };

              nativeBuildInputs = [
                pkgs.cmake
                protobuf-pkg
                protobuf-plugins
                codegen-tools
                pkgs.ruby
                pkgs.which
              ];

              buildInputs = [
                pkgs.boost
                pkgs.nlohmann_json
                rustLib
              ];

              # Let cmake know where the pre-built Rust static lib lives
              cmakeFlags = [
                "-DCMAKE_BUILD_TYPE=Release"
                "-DCMAKE_C_COMPILER=${stdenv.cc}/bin/clang"
                "-DCMAKE_CXX_COMPILER=${stdenv.cc}/bin/clang++"
                "-DTW_UNIT_TESTS=OFF"
                "-DBUILD_TESTING=OFF"
                "-DBoost_INCLUDE_DIR=${pkgs.boost}/include"
                "-DWALLET_CORE_RS_TARGET_DIR=${rustLib}"
              ];

              postPatch = ''
                # Point cmake's Protobuf search at the Nix-provided installation
                substituteInPlace cmake/Protobuf.cmake \
                  --replace-fail \
                    'set(protobuf_SOURCE_DIR ''${CMAKE_CURRENT_LIST_DIR}/../build/local/src/protobuf/protobuf-3.20.3)' \
                    'set(protobuf_SOURCE_DIR ${protobuf-pkg}/include)' \
                  --replace-fail \
                    'set(protobuf_source_dir ''${CMAKE_CURRENT_LIST_DIR}/../build/local/src/protobuf/protobuf-3.20.3)' \
                    'set(protobuf_source_dir ${protobuf-pkg}/include)'

                # Generate .pb.cc / .pb.h from proto files (runs in the build tree)
                for proto in src/proto/*.proto; do
                  protoc -I=src/proto --cpp_out=src/proto "$proto"
                done
                for dir in src/Tron/Protobuf src/Zilliqa/Protobuf src/Hedera/Protobuf; do
                  [ -d "$dir" ] && protoc -I="$dir" --cpp_out="$dir" "$dir"/*.proto
                done
              '';

              preConfigure = ''
                # Sanity-check that the Rust static library was actually built
                if [ ! -f "${rustLib}/lib/libwallet_core_rs.a" ]; then
                  echo "ERROR: Rust library not found at ${rustLib}/lib/libwallet_core_rs.a"
                  ls "${rustLib}/lib/" || true
                  exit 1
                fi
              '';

              buildPhase = ''
                runHook preBuild
                make -C "$cmakeBuildDir" -j"$NIX_BUILD_CORES" TrustWalletCore
                runHook postBuild
              '';

              installPhase = ''
                runHook preInstall
                install -Dm644 \
                  "$(find "$cmakeBuildDir" -name 'libTrustWalletCore.a' -print -quit)" \
                  "$out/lib/libTrustWalletCore.a"
                cp -r include "$out/include"
                find src -name '*.pb.h' -exec install -Dm644 {} "$out/include/" \;
                runHook postInstall
              '';

              meta = {
                description = "Trust Wallet Core FFI library";
                license     = lib.licenses.asl20;
                platforms   = lib.platforms.linux;
              };
            };

        in {
          nci.projects."wallet-core-rs" = {
            path   = ./rust;
            export = true;

            drvConfig.mkDerivation = {
              # Include src/proto/ so tw_proto's build.rs finds its proto files
              src = lib.cleanSourceWith {
                src    = ./.;
                filter = rustSourceFilter;
              };

              nativeBuildInputs = [ protobuf-pkg ];

              postInstall = ''
                # Crane installs the rlib; also expose the staticlib
                if [ -f target/release/libwallet_core_rs.a ]; then
                  install -Dm644 target/release/libwallet_core_rs.a \
                    "$out/lib/libwallet_core_rs.a"
                fi
              '';
            };
          };

          packages = {
            wallet-core-rs  = config.nci.outputs."wallet-core-rs".packages.release;
            wallet-core-ffi = mkWalletCoreFFI
              config.nci.outputs."wallet-core-rs".packages.release;
            default         = mkWalletCoreFFI
              config.nci.outputs."wallet-core-rs".packages.release;
          };

          devShells = {
            default =
              config.nci.outputs."wallet-core-rs".devShell.overrideAttrs (old: {
                name = "wallet-core-dev";
                nativeBuildInputs = (old.nativeBuildInputs or []) ++ [
                  pkgs.cmake
                  pkgs.boost
                  pkgs.nlohmann_json
                  pkgs.ruby
                  pkgs.go
                  pkgs.which
                  protobuf-pkg
                  protobuf-plugins
                  codegen-tools
                ];
                shellHook = ''
                  ${old.shellHook or ""}
                  export CC="${stdenv.cc}/bin/clang"
                  export CXX="${stdenv.cc}/bin/clang++"
                  export PATH="${protobuf-pkg}/bin:${protobuf-plugins}/bin:${codegen-tools}/bin:$PATH"
                  echo "wallet-core dev shell — clang $(clang --version | head -1)"
                  echo "  nix build .#wallet-core-rs   # Rust static lib"
                  echo "  nix build .#wallet-core-ffi  # full C++ FFI"
                '';
              });

            rust = config.nci.outputs."wallet-core-rs".devShell;
          };

          checks = {
            wallet-core-rs-build  = config.nci.outputs."wallet-core-rs".packages.release;
            wallet-core-ffi-build = mkWalletCoreFFI
              config.nci.outputs."wallet-core-rs".packages.release;
          };
        };
    };
}
