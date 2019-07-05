package topology

import "testing"

func TestNodeGroupDiff(t *testing.T) {
	source := newNodeGroup()
	source.put(&NodeInfo{RunID: "pufan"})
	source.put(&NodeInfo{RunID: "jiaming"})

	new := newNodeGroup()
	new.put(&NodeInfo{RunID: "ruibing"})

	incr, _, hasdiff := source.diff(new)
	if !hasdiff {
		t.Fatal("expected has differ")
	}

	if incr[0].RunID != "ruibing" {
		t.Fatal("expected incr")
	}

	new1 := newNodeGroup()
	new1.put(&NodeInfo{RunID: "pufan"})
	new1.put(&NodeInfo{RunID: "biaoyun"})

	incr, decr, hasdiff := source.diff(new1)
	if !hasdiff {
		t.Fatal("expected has differ")
	}
	if incr[0].RunID != "biaoyun" {
		t.Fatal("expected incr")
	}
	if len(decr) < 1 || decr[0].RunID != "jiaming" {
		t.Fatal("expected decr")
	}

}
