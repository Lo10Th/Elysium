class Ely < Formula
  desc "Elysium CLI - The API App Store"
  homepage "https://github.com/Lo10Th/Elysium"
  version "0.2.0"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/Lo10Th/Elysium/releases/download/v#{version}/ely-darwin-amd64.tar.gz"
      sha256 "DARWIN_AMD64_SHA256_PLACEHOLDER"
    end

    on_arm do
      url "https://github.com/Lo10Th/Elysium/releases/download/v#{version}/ely-darwin-arm64.tar.gz"
      sha256 "DARWIN_ARM64_SHA256_PLACEHOLDER"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/Lo10Th/Elysium/releases/download/v#{version}/ely-linux-amd64.tar.gz"
      sha256 "LINUX_AMD64_SHA256_PLACEHOLDER"
    end

    on_arm do
      url "https://github.com/Lo10Th/Elysium/releases/download/v#{version}/ely-linux-arm64.tar.gz"
      sha256 "LINUX_ARM64_SHA256_PLACEHOLDER"
    end
  end

  def install
    bin.install "ely"
  end

  test do
    output = shell_output("#{bin}/ely --version")
    assert_match "Elysium", output
    assert_match version.to_s, output
  end
end
