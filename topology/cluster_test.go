package topology

import "testing"

func TestNodeGroupDiff(t *testing.T) {
	source := newNodeGroup()
	source.put(&NodeInfo{Id: "pufan"})
	source.put(&NodeInfo{Id: "jiaming"})

	new := newNodeGroup()
	new.put(&NodeInfo{Id: "ruibing"})

	incr, _, hasdiff := source.diff(new)
	if !hasdiff {
		t.Fatal("expected has differ")
	}

	if incr[0].Id != "ruibing" {
		t.Fatal("expected incr")
	}

	new1 := newNodeGroup()
	new1.put(&NodeInfo{Id: "pufan"})
	new1.put(&NodeInfo{Id: "biaoyun"})

	incr, decr, hasdiff := source.diff(new1)
	if !hasdiff {
		t.Fatal("expected has differ")
	}
	if incr[0].Id != "biaoyun" {
		t.Fatal("expected incr")
	}
	if len(decr) < 1 || decr[0].Id != "jiaming" {
		t.Fatal("expected decr")
	}

}
