#!/bin/bash
# Release script for MCP Vikunja
# Builds binaries for multiple architectures and operating systems

set -e

# Configuration
TARGET_ARCHS="amd64 arm64"
TARGET_OS="linux darwin"
RELEASE_DIR="release"

echo "🚀 Starting release build process..."

# Clean previous builds
rm -rf ${RELEASE_DIR}
mkdir -p ${RELEASE_DIR}

# Function to compile for specific platform
function compile() {
  local name=$1
  local arch=$2
  local os=$3
  local output="${RELEASE_DIR}/${arch}_${os}/${name}"

  echo "  Building ${name} for ${arch}/${os}..."

  CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o "${output}" \
    "./cmd/${name}"

  # Verify binary was created
  if [ ! -f "${output}" ]; then
    echo "❌ Failed to build ${name} for ${arch}/${os}"
    exit 1
  fi

  echo "  ✓ Built: ${output}"
}

# Build all combinations
for arch in ${TARGET_ARCHS}; do
  for os in ${TARGET_OS}; do
    echo ""
    echo "📦 Building for ${arch}/${os}..."
    mkdir -p "${RELEASE_DIR}/${arch}_${os}"

    # Build mcp-vikunja server
    compile "mcp-vikunja" "${arch}" "${os}"

    # Build vikunja-cli client
    compile "vikunja-cli" "${arch}" "${os}"

    # Create tarball
    (cd "${RELEASE_DIR}/${arch}_${os}"
      tar zcvf "../mcp-vikunja_${arch}_${os}.tgz" *
    )

    echo "✓ Created: ${RELEASE_DIR}/mcp-vikunja_${arch}_${os}.tgz"
  done
done

# Generate checksums
echo ""
echo "🔐 Generating checksums..."
cd ${RELEASE_DIR}
sha256sum *.tgz > checksums.txt
cd ..
echo "✓ Checksums written to ${RELEASE_DIR}/checksums.txt"

echo ""
echo "✅ Release build complete!"
echo ""
echo "Artifacts:"
ls -lh ${RELEASE_DIR}/*.tgz
echo ""
echo "Checksums:"
cat ${RELEASE_DIR}/checksums.txt
