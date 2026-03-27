{
  description = "Trust Wallet Core - C++ FFI with Rust core";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
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
          protobuf = pkgs.protobuf_21;

          protobuf-plugins = stdenv.mkDerivation {
            pname = "wallet-core-protobuf-plugins";
            version = "0.1.0";
            src = ./protobuf-plugin;
            nativeBuildInputs = [ pkgs.cmake protobuf ];
            cmakeFlags = [ "-DProtobuf_PROTOC_EXECUTABLE=${protobuf}/bin/protoc" ];
          };

          codegen-v2-tool = config.nci.outputs."codegen-v2".packages.release;

          wallet-core-generated = { rustLib }:
            pkgs.runCommand "wallet-core-generated"
              { nativeBuildInputs = [ codegen-v2-tool ]; }
              ''
                mkdir -p $TMPDIR/workspace/{rust/bindings,include/TrustWalletCore,src/Generated,codegen-v2/manifest}
                cp ${rustLib}/bindings/*.yaml $TMPDIR/workspace/rust/bindings/
                cp ${./codegen-v2/manifest}/*.yaml $TMPDIR/workspace/codegen-v2/manifest/
                (cd $TMPDIR/workspace/codegen-v2 && parser cpp)
                mkdir -p $out/{include/TrustWalletCore,src/Generated}
                cp $TMPDIR/workspace/include/TrustWalletCore/*.h $out/include/TrustWalletCore/
                 cp $TMPDIR/workspace/src/Generated/* $out/src/Generated/
              '';

          wallet-core-ffi = { rustLib, generated }:
            stdenv.mkDerivation {
              pname = "wallet-core-ffi";
              version = "0.1.0";
              src = ./.;

              nativeBuildInputs = [
                pkgs.cmake protobuf protobuf-plugins
                pkgs.ruby pkgs.which
              ];

              buildInputs = [ pkgs.boost pkgs.nlohmann_json protobuf ];

              cmakeFlags = [
                "-DCMAKE_BUILD_TYPE=Release"
                "-DTW_UNIT_TESTS=OFF"
                "-DBUILD_TESTING=OFF"
                "-DBoost_INCLUDE_DIR=${pkgs.boost}/include"
                "-DWALLET_CORE_RS_TARGET_DIR=${rustLib}"
              ];

              postPatch = ''
                mkdir -p include/TrustWalletCore src/Generated src/rust/bindgen
                cp ${generated}/include/TrustWalletCore/* include/TrustWalletCore/
                cp ${generated}/src/Generated/* src/Generated/
                ln -sf ${rustLib}/include/WalletCoreRSBindgen.h src/rust/bindgen/

                patchShebangs tools/ codegen/bin/

                mkdir -p build/local/{bin,lib,include}
                ln -sf ${protobuf}/bin/protoc build/local/bin/
                ln -sf ${protobuf}/lib/libprotobuf.a build/local/lib/
                ln -sf ${protobuf}/include/google build/local/include/
                ln -sf ${protobuf-plugins}/bin/protoc-gen-c-typedef build/local/bin/
                ln -sf ${protobuf-plugins}/bin/protoc-gen-swift-typealias build/local/bin/

                export PREFIX=$PWD/build/local
                export PATH="$PREFIX/bin:$PATH"
                substituteInPlace tools/generate-files \
                  --replace-fail 'tools/rust-bindgen "$@"' ': # rust-bindgen skipped'
                tools/generate-files native

                cat > cmake/Protobuf.cmake << 'EOF'
                find_package(Protobuf REQUIRED)
                add_library(protobuf INTERFACE)
                target_link_libraries(protobuf INTERFACE protobuf::libprotobuf)
                target_include_directories(protobuf INTERFACE ''${Protobuf_INCLUDE_DIRS})
                EOF
              '';

              buildPhase = ''
                make -j"$NIX_BUILD_CORES" TrustWalletCore
              '';

              installPhase = ''
                install -Dm644 libTrustWalletCore.a $out/lib/libTrustWalletCore.a
                install -Dm644 trezor-crypto/libTrezorCrypto.a $out/lib/libTrezorCrypto.a
                install -Dm644 ${rustLib}/lib/libwallet_core_rs.a $out/lib/libwallet_core_rs.a
                cp -r ../include $out/include
                find $src/src -name '*.pb.h' -exec install -Dm644 {} $out/include/ \;
              '';
            };

        in {
          nci.projects = {
            wallet-core-rs = {
              path = ./rust;
              export = true;
              drvConfig.mkDerivation = {
                nativeBuildInputs = [ protobuf pkgs.rust-cbindgen ];
                preBuild = ''
                  export CARGO_WORKSPACE_DIR="$out"
                  mkdir -p $out/bindings ../src/
                  ln -s ${./src/proto} ../src/proto
                  cp ${./registry.json} ../registry.json
                '';
                postInstall = ''
                  cbindgen --crate wallet-core-rs --output $out/include/WalletCoreRSBindgen.h ${./rust}
                '';
              };
            };

            codegen-v2 = {
              path = ./codegen-v2;
              export = false;
              drvConfig.mkDerivation.nativeBuildInputs = [ protobuf ];
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
              pkgs.ruby pkgs.which protobuf protobuf-plugins
              codegen-v2-tool
              pkgs.rustc pkgs.cargo pkgs.clippy pkgs.rustfmt
            ];
            shellHook = ''
              export CC="${stdenv.cc}/bin/clang"
              export CXX="${stdenv.cc}/bin/clang++"
            '';
          };

          devShells.go = let
            ffi = config.packages.wallet-core-ffi;
          in pkgs.mkShell {
            nativeBuildInputs = [ pkgs.go ];
            buildInputs = [ protobuf ];
            shellHook = ''
              export CGO_CFLAGS="-I${ffi}/include"
              export CGO_LDFLAGS="-L${ffi}/lib -L${protobuf}/lib -Wl,-rpath,${protobuf}/lib -lTrustWalletCore -lwallet_core_rs -lprotobuf -lTrezorCrypto -lstdc++ -lm"
              export LD_LIBRARY_PATH="${protobuf}/lib:$LD_LIBRARY_PATH"
            '';
          };
        };
    };
}
