class Ghostctl < Formula
  desc "CLI tool for managing ephemeral Kubernetes clusters using vCluster"
  homepage "https://github.com/ghostcluster-ai/ghostctl"
  url "https://github.com/ghostcluster-ai/ghostctl/archive/refs/tags/v1.0.3.tar.gz"
  sha256 "c661268b970d6e20478a702a1a009fb0a1d70d0f526a836621b3bdf60df6ecc8"
  version "1.0.3"
  license "MIT"

  depends_on "go" => :build
  depends_on "kubectl"
  depends_on "vcluster"

  def install
    # Build the binary
    system "make", "build"
    
    # Install to bin directory
    bin.install "bin/ghostctl"
  end

  test do
    system "#{bin}/ghostctl", "--help"
  end
end
