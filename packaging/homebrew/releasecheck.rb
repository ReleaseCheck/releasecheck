class Releasecheck < Formula
  desc "Deterministic release-integrity evidence for npm and PyPI packages"
  homepage "https://github.com/ReleaseCheck/releasecheck"
  version "0.1.2"
  license "Apache-2.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/ReleaseCheck/releasecheck/releases/download/v0.1.2/releasecheck-v0.1.2-darwin-arm64.tar.gz"
      sha256 "a23cef31e1687dd9c61caa4813dbb8d7670b21890abfc54cd86a99dc1fa72838"
    else
      url "https://github.com/ReleaseCheck/releasecheck/releases/download/v0.1.2/releasecheck-v0.1.2-darwin-amd64.tar.gz"
      sha256 "f4d85808a58faa2300fb77be5c5f651c88699de1e4df81599bb84062ff2c0505"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/ReleaseCheck/releasecheck/releases/download/v0.1.2/releasecheck-v0.1.2-linux-arm64.tar.gz"
      sha256 "1dcc0b6b74c733681649b80e64b0bfdd27ca0dc0019fbc2171bbb3b370cfa281"
    else
      url "https://github.com/ReleaseCheck/releasecheck/releases/download/v0.1.2/releasecheck-v0.1.2-linux-amd64.tar.gz"
      sha256 "6816c31d641f2467fa3c12f37f16c0861e8de99492a3f8275ebb0676142a6947"
    end
  end

  def install
    bin.install "releasecheck"
  end

  test do
    assert_match "releasecheck", shell_output("#{bin}/releasecheck --version")
  end
end
