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
               # Replace /usr/bin/env ruby with nix ruby
               substituteInPlace "$out/bin/$(basename "$bin")" \
                 --replace-fail '#!/usr/bin/env ruby' '#!${pkgs.ruby}/bin/ruby'
               wrapProgram "$out/bin/$(basename "$bin")" \
                 --prefix PATH : ${lib.makeBinPath [ pkgs.ruby pkgs.which ]}
             done
           '';

          # Rust build needs: rust/ workspace + src/proto/ (for tw_proto)
          rustSourceFilter = path: type:
            let
              rel = lib.removePrefix (toString ./.) (toString path);
              isProtoDir = rel == "/src" || rel == "/src/proto" || lib.hasPrefix "/src/proto/" rel;
              isRustDir = rel == "/rust" || lib.hasPrefix "/rust/" rel;
              isRegistry = rel == "/registry.json";
            in
              isRustDir || isProtoDir || isRegistry || rel == "";

           # C++ build needs headers, sources, cmake, and generated proto files
            cppSourceFilter = path: type:
              let
                rel      = lib.removePrefix (toString ./.) (toString path);
                extMatch = ext: lib.hasSuffix ext (baseNameOf path);
              in
                type == "directory" ||
                extMatch ".cpp" || extMatch ".cc"  || extMatch ".c"    ||
                extMatch ".h"   || extMatch ".hpp" || extMatch ".proto"  ||
                extMatch ".cmake" || extMatch ".txt" || extMatch ".json" ||
                extMatch ".in" || extMatch ".md" ||
                 lib.hasPrefix "/src/"          rel ||
                  lib.hasPrefix "/include/"      rel ||
                  lib.hasPrefix "/trezor-crypto/" rel ||
                  lib.hasPrefix "/cmake/"        rel ||
                  lib.hasPrefix "/jni/cpp/"      rel ||
                  lib.hasPrefix "/swift/"        rel ||
                  lib.hasPrefix "/wasm/"         rel ||
                  lib.hasPrefix "/codegen/"      rel ||
                  lib.hasPrefix "/codegen-v2/"   rel ||
                  lib.hasPrefix "/tools/"        rel ||
                  lib.hasPrefix "/rust/bindings/" rel ||
                  rel == "/registry.json";

           bindgenHeader = pkgs.writeText "WalletCoreRSBindgen.h" (builtins.readFile ./src/rust/bindgen/WalletCoreRSBindgen.h);

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
                 protobuf-pkg
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
                   "-DProtobuf_ROOT=${protobuf-pkg}"
                   "-DProtobuf_INCLUDE_DIR=${protobuf-pkg}/include"
                   "-DProtobuf_LIBRARY=${protobuf-pkg}/lib/libprotobuf.a"
                 ];

                postPatch = ''
                    # Copy the pre-generated Rust bindgen header
                    mkdir -p src/rust/bindgen
                    cp ${bindgenHeader} src/rust/bindgen/WalletCoreRSBindgen.h

                    # Fix shebang in generate-files script
                    patchShebangs tools/generate-files
                    patchShebangs tools/parse_args
                    patchShebangs tools/rust-bindgen
                    patchShebangs tools/doxygen_convert_comments
                    patchShebangs codegen/bin

                    # Set up PREFIX directory structure that tools/generate-files expects
                    mkdir -p $PWD/build/local/bin $PWD/build/local/lib $PWD/build/local/include
                    ln -sf ${protobuf-pkg}/bin/protoc $PWD/build/local/bin/protoc
                    ln -sf ${protobuf-pkg}/lib/libprotobuf.a $PWD/build/local/lib/libprotobuf.a
                    ln -sf ${protobuf-pkg}/include/google $PWD/build/local/include/google
                    ln -sf ${protobuf-plugins}/bin/protoc-gen-c-typedef $PWD/build/local/bin/protoc-gen-c-typedef
                    ln -sf ${protobuf-plugins}/bin/protoc-gen-swift-typealias $PWD/build/local/bin/protoc-gen-swift-typealias

                    # Run the official code generation script (skip rust-bindgen since we have pre-built Rust lib)
                    export PREFIX=$PWD/build/local
                    export PATH="$PREFIX/bin:$PATH"
                    export LD_LIBRARY_PATH="$PREFIX/lib:$LD_LIBRARY_PATH"
                    # Patch out rust-bindgen call since Rust lib is already built
                    substituteInPlace tools/generate-files \
                      --replace-fail 'tools/rust-bindgen "$@"' '# rust-bindgen skipped - using pre-built Rust library'
                    tools/generate-files native

                    # Replace the entire cmake/Protobuf.cmake to use nix-provided protobuf
                    cat > cmake/Protobuf.cmake << 'EOF'
                    # Use nix-provided protobuf
                    find_package(Protobuf REQUIRED)
                    add_library(protobuf INTERFACE)
                    target_link_libraries(protobuf INTERFACE protobuf::libprotobuf)
                    target_include_directories(protobuf INTERFACE ''${Protobuf_INCLUDE_DIRS})
                    EOF
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
                 make -j"$NIX_BUILD_CORES" TrustWalletCore
                 runHook postBuild
               '';

               installPhase = ''
                 runHook preInstall
                 # Find the built library - cmake may use different build directories
                 LIB_PATH=$(find . -name 'libTrustWalletCore.a' -print -quit)
                 if [ -z "$LIB_PATH" ]; then
                   echo "Error: Could not find libTrustWalletCore.a"
                   find . -name '*.a' | head -20
                   exit 1
                 fi
                 install -Dm644 "$LIB_PATH" "$out/lib/libTrustWalletCore.a"
                 # Copy include directory from source
                 cp -r "$src/include" "$out/include"
                 find "$src/src" -name '*.pb.h' -exec install -Dm644 {} "$out/include/" \;
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

               preBuild = ''
                 mkdir -p $PWD/../src
                 cp -r ${./src/proto} $PWD/../src/proto
                 chmod -R u+w $PWD/../src
                 export WALLET_CORE_PROTO_DIR=$PWD/../src/proto
                 cp ${./registry.json} $PWD/../registry.json
               '';

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
