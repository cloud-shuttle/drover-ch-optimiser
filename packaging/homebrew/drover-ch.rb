# Optional Homebrew formula — install from a tap or use:
#   brew install ./packaging/homebrew/drover-ch.rb
#
# After tagging v0.1.0, update sha256 from the release checksums.txt.

class DroverCh < Formula
  desc "ClickHouse warehouse auditor (Drover family)"
  homepage "https://github.com/cloud-shuttle/drover-ch-optimiser"
  url "https://github.com/cloud-shuttle/drover-ch-optimiser/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "REPLACE_WITH_RELEASE_SHA256"
  license "Apache-2.0"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", bin/"drover-ch", "./cmd/drover-ch"
  end

  test do
    assert_match "drover-ch", shell_output("#{bin}/drover-ch --help")
    assert_match "insert-mutation", shell_output("#{bin}/drover-ch rules")
  end
end
