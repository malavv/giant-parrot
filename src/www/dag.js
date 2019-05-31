const unique = arr => [...new Set(arr)];
const emptyCallback = () => {};
class Queue {
    constructor() { this.data = []; }
    enqueue(o) { this.data.unshift(o); }
    dequeue() { return this.data.pop(); }
    size() { return this.data.length; }
}

export class DAG {
    constructor(idFun) {
        this.id2idx = new Map();
        this.idFun = idFun;
        this.nodes = [];
        this.links = [];
        this.len = new Map();
    }

    /**
     * Adds node if missing
     * @param node
     * @returns {boolean} True if its new.
     */
    addNode(node) {
        const id = this.idFun(node);
        if (this.id2idx.has(id))
            return false; /* not new */
        this.id2idx.set(id, this.nodes.length);
        this.nodes.push(node);
        return true; /* is new */
    }

    hasLink(src, dst) {
        const
            sid = this.idFun(src),
            did = this.idFun(dst);

        return this.links.some(lk => nodeID(lk.source) === sid && nodeID(lk.target) === did);
    }

    addLink(srcNode, dstNode) {
        const srcID = this.idFun(srcNode);
        const dstID = this.idFun(dstNode);

        if (this.hasLink(srcNode, dstNode))
            return;

        // get from the known nodes instead of just entering them.
        this.links.push({
            source: this.nodes[this.id2idx.get(srcID)],
            target: this.nodes[this.id2idx.get(dstID)]
        });

        // Update path length
        this.len.set([srcID, dstID].join("~"), 1);
        for (const [key, value] of this.len.entries()) {
            const [sid, did] = key.split('~');
            if (did === srcID)
                this.len.set([sid, dstID].join("~"), value + 1);
        }
    }

    pruneNode(node) {
        const id = this.idFun(node);
        const idx = this.id2idx.get(id);
        // Remove Node
        this.nodes.splice(idx, 1);
        // Adjust Idx Map
        this.nodes.map(n => this.idFun(n))
            .forEach((nid, nidx) => { this.id2idx.set(nid, nidx); });
        // Remove links?
        this.links = this.links.filter(lk => this.idFun(lk.source) !== id);
        this.links = this.links.filter(lk => this.idFun(lk.target) !== id);
    }

    rmLinksWithSourceID(id) {
        // fix that the links are moved to the node themselves as soon as the graph loads.
        if (this.links[0].source.id === undefined)
            return; // Prior to the graph loading.

        // Actually removing the links
        const toBeRemoved = this.links.filter(l => this.idFun(l.source) === id);
        this.links = this.links.filter(l => this.idFun(l.source) !== id);

        // Investigate disconnected nodes for possible separate forests.
        const nodesInSepForest = new Set();
        for (let lk of toBeRemoved) {
            if (isFinite(this.shortestPath(lk.source, lk.target)))
                continue;
            // Here the target is part of a separate forest;
            this.bfs(lk.target, id => { nodesInSepForest.add(id); });
        }

        for (let n of nodesInSepForest) {
            this.pruneNode(this.nodes[this.id2idx.get(n)]);
        }
    }

    pathLength(srcID, dstID) {
        // Root
        if (srcID === dstID)
            return 0;

        // Cached
        let len = this.len.get([srcID, dstID].join("~"));
        if (len != null)
            return len;

        return 5;
    }

    _adj(sID) {
        return this.links.filter(lk => this.idFun(lk.source) === sID).map(lk => this.idFun(lk.target));
    }

    shortestPath(src, dst) {
        let shortest = Number.POSITIVE_INFINITY;
        this.bfs(src, (id, dist) => {
            if (id === this.idFun(dst))
                shortest = dist;
        });
        return shortest;
    }

    bfs(center, callback = emptyCallback) {
        const colors = new Map();
        const dists = new Map();
        const parents = new Map();

        const s = this.idFun(center);
        for (let u of this.nodes.map(n => this.idFun(n))) {
            colors.set(u, "white");
            dists.set(u, Number.POSITIVE_INFINITY);
            parents.set(u, null);
        }

        colors.set(s, "gray");
        dists.set(s, 0);
        parents.set(s, null);

        const q = new Queue();
        q.enqueue(s);

        while (q.size() > 0) {
            let u = q.dequeue();
            for (let v of this._adj(u)) {
                if (colors.get(v) !== "white")
                    continue;
                colors.set(v, "gray");
                dists.set(v, u.d + 1);
                parents.set(v, u);
                q.enqueue(v);
            }
            colors.set(u, "gray");
            callback(u, dists.get(u), parents.get(u) ? parents.get(u) : null);
        }
    }
}

console.info("DAG.js is loaded");