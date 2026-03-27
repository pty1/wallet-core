{
  description = "Trust Wallet Core - C++ FFI with Rust core";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    nixpkgs-protobuf.url = "github:nixos/nixpkgs/nixos-23.11";
    flake-parts.url = "github:hercules-ci/flake-parts";
    nci.url = "github:yusdacra/nix-cargo-integration";
  };

  outputs = inputs:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.nci.flakeModule ];

      systems = [ "x86_64-linux" "aarch64-linux" ];

      perSystem = { config, pkgs, system, ... }:
        let
          inherit (pkgs) lib;
          stdenv = pkgs.llvmPackages.stdenv;

          # Protobuf 3.20 for compatibility
          protobuf-pkg = inputs.nixpkgs-protobuf.legacyPackages.${system}.protobuf3_20;

          # Protobuf codegen plugins (C++ only)
          protobuf-plugins = stdenv.mkDerivation {
            pname = "wallet-core-protobuf-plugins";
            version = "0.1.0";
            src = ./protobuf-plugin;

            nativeBuildInputs = [ pkgs.cmake protobuf-pkg ];

            postPatch = ''
              substituteInPlace CMakeLists.txt \
                --replace-fail 'find_package(Protobuf CONFIG REQUIRED PATH ''${PREFIX}/lib/pkgconfig)' \
                               'find_package(Protobuf REQUIRED)'
            '';

            cmakeFlags = [
              "-DProtobuf_PROTOC_EXECUTABLE=${protobuf-pkg}/bin/protoc"
            ];
          };

          # Ruby codegen tools (legacy, for protobuf generation)
          codegen-tools = pkgs.runCommand "wallet-core-codegen" {
            nativeBuildInputs = [ pkgs.makeWrapper ];
          } ''
            mkdir -p $out/bin
            for bin in ${./codegen}/bin/*; do
              install -m755 "$bin" "$out/bin/$(basename "$bin")"
              substituteInPlace "$out/bin/$(basename "$bin")" \
                --replace-fail '#!/usr/bin/env ruby' '#!${pkgs.ruby}/bin/ruby'
              wrapProgram "$out/bin/$(basename "$bin")" \
                --prefix PATH : ${lib.makeBinPath [ pkgs.ruby pkgs.which ]}
            done
          '';

          # Binary tools from nci
          codegen-v2-tool = config.nci.outputs."codegen-v2".packages.release;

          # Run codegen-v2 to produce C++ headers/sources from YAML bindings
          # The tool has hardcoded paths, so we create the expected workspace structure
          wallet-core-generated = { rustLib }:
            pkgs.runCommand "wallet-core-generated"
              {
                nativeBuildInputs = [ codegen-v2-tool ];
              }
              ''
                # codegen-v2 expects to be run from a directory named 'codegen-v2'
                # with sibling directories: rust/bindings/, include/TrustWalletCore/, src/Generated/
                mkdir -p $TMPDIR/workspace/rust/bindings
                mkdir -p $TMPDIR/workspace/include/TrustWalletCore
                mkdir -p $TMPDIR/workspace/src/Generated
                mkdir -p $TMPDIR/workspace/codegen-v2

                # Copy YAML bindings from the Rust lib output
                cp ${rustLib}/bindings/*.yaml $TMPDIR/workspace/rust/bindings/ 2>/dev/null || {
                  echo "ERROR: No YAML bindings found in ${rustLib}/bindings/"
                  exit 1
                }

                echo "Generating C++ from $(ls $TMPDIR/workspace/rust/bindings/*.yaml | wc -l) YAML binding files..."

                # Run codegen-v2 cpp from the expected directory
                (cd $TMPDIR/workspace/codegen-v2 && parser cpp)

                # Copy generated files to output
                mkdir -p $out/include/TrustWalletCore $out/src/Generated
                cp $TMPDIR/workspace/include/TrustWalletCore/*.h $out/include/TrustWalletCore/
                cp $TMPDIR/workspace/src/Generated/*.cpp $out/src/Generated/ 2>/dev/null || true
                cp $TMPDIR/workspace/src/Generated/*.h $out/src/Generated/ 2>/dev/null || true

                echo "Generated files:"
                find $out -type f
              '';

          # Main C++ FFI library
          wallet-core-ffi = { rustLib, generated }:
            stdenv.mkDerivation {
              pname = "wallet-core-ffi";
              version = "0.1.0";
              src = ./.;

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
              ];

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
                # Copy generated C++ headers/sources from codegen-v2
                mkdir -p include/TrustWalletCore src/Generated
                cp ${generated}/include/TrustWalletCore/* include/TrustWalletCore/
                cp ${generated}/src/Generated/* src/Generated/ 2>/dev/null || true

                # Link Rust bindgen header
                mkdir -p src/rust/bindgen
                ln -sf ${rustLib}/include/WalletCoreRSBindgen.h src/rust/bindgen/WalletCoreRSBindgen.h

                # Set up PREFIX directory for tools/generate-files
                patchShebangs tools/generate-files tools/parse_args tools/doxygen_convert_comments codegen/bin

                mkdir -p $PWD/build/local/bin $PWD/build/local/lib $PWD/build/local/include
                ln -sf ${protobuf-pkg}/bin/protoc $PWD/build/local/bin/protoc
                ln -sf ${protobuf-pkg}/lib/libprotobuf.a $PWD/build/local/lib/libprotobuf.a
                ln -sf ${protobuf-pkg}/include/google $PWD/build/local/include/google
                ln -sf ${protobuf-plugins}/bin/protoc-gen-c-typedef $PWD/build/local/bin/protoc-gen-c-typedef
                ln -sf ${protobuf-plugins}/bin/protoc-gen-swift-typealias $PWD/build/local/bin/protoc-gen-swift-typealias

                # Run code generation (skipping rust-bindgen since we have pre-built Rust)
                export PREFIX=$PWD/build/local
                export PATH="$PREFIX/bin:$PATH"
                substituteInPlace tools/generate-files \
                  --replace-fail 'tools/rust-bindgen "$@"' ': # rust-bindgen skipped'
                tools/generate-files native

                # Use nix-provided protobuf in cmake
                cat > cmake/Protobuf.cmake << 'EOF'
                find_package(Protobuf REQUIRED)
                add_library(protobuf INTERFACE)
                target_link_libraries(protobuf INTERFACE protobuf::libprotobuf)
                target_include_directories(protobuf INTERFACE ''${Protobuf_INCLUDE_DIRS})
                EOF
              '';

              preConfigure = ''
                if [ ! -f "${rustLib}/lib/libwallet_core_rs.a" ]; then
                  echo "ERROR: Rust library not found"
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
                find . -name 'libTrustWalletCore.a' -exec install -Dm644 {} "$out/lib/libTrustWalletCore.a" \;
                cp -r "$src/include" "$out/"
                find "$src/src" -name '*.pb.h' -exec install -Dm644 {} "$out/include/" \;
                runHook postInstall
              '';
            };

        in {
          # NCI projects for Rust builds
          nci.projects = {
            wallet-core-rs = {
              path = ./rust;
              export = true;
              drvConfig.mkDerivation = {
                nativeBuildInputs = [ protobuf-pkg pkgs.rust-cbindgen ];

                preBuild = ''
                  # Set workspace dir so tw_ffi macro writes bindings to output
                  export CARGO_WORKSPACE_DIR="$out"
                  mkdir -p $out/bindings
                  mkdir -p ../src/
                  ln -s ${./src/proto} ../src/proto
                  cp ${./registry.json} ../registry.json
                '';

                postInstall = ''
                  # Generate cbindgen header
                  cbindgen --crate wallet-core-rs --output $out/include/WalletCoreRSBindgen.h ${./rust}

                  # Verify bindings were created
                  if [ ! -f "$out/bindings"/*.yaml 2>/dev/null ]; then
                    echo "WARNING: No YAML bindings found in $out/bindings/"
                    ls -la $out/bindings/ || true
                  else
                    echo "YAML bindings: $(ls $out/bindings/*.yaml | wc -l) files"
                  fi
                '';
              };
            };

            codegen-v2 = {
              path = ./codegen-v2;
              export = false;
              drvConfig.mkDerivation = {
                nativeBuildInputs = [ protobuf-pkg ];
              };
            };
          };

          packages = {
            wallet-core-rs = config.nci.outputs."wallet-core-rs".packages.release;
            codegen-v2 = codegen-v2-tool;
            wallet-core-generated = wallet-core-generated {
              rustLib = config.nci.outputs."wallet-core-rs".packages.release;
            };
            wallet-core-ffi = wallet-core-ffi {
              rustLib = config.nci.outputs."wallet-core-rs".packages.release;
              generated = wallet-core-generated {
                rustLib = config.nci.outputs."wallet-core-rs".packages.release;
              };
            };
            default = config.packages.wallet-core-ffi;
          };

          devShells.default = pkgs.mkShell {
            nativeBuildInputs = [
              pkgs.cmake pkgs.boost pkgs.nlohmann_json
              pkgs.ruby pkgs.which protobuf-pkg protobuf-plugins
              codegen-tools codegen-v2-tool
              pkgs.rustc pkgs.cargo pkgs.clippy pkgs.rustfmt
            ];
            shellHook = ''
              export CC="${stdenv.cc}/bin/clang"
              export CXX="${stdenv.cc}/bin/clang++"
              echo "wallet-core dev shell"
            '';
          };
        };
    };
}
