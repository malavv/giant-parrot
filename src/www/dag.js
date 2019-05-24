export class DAG {
    constructor(idFun) {
        this.id2idx = new Map();
        this.idFun = idFun;
        this.nodes = [];
        this.links = [];
        this.len = new Map();
    }

    addNode(node) {
        const id = this.idFun(node);
        if (this.id2idx.has(id))
            return false; /* not new */
        this.id2idx.set(id, this.nodes.length);
        this.nodes.push(node);
        return true; /* is new */
    }

    addLink(srcNode, dstNode) {
        const srcID = this.idFun(srcNode);
        const dstID = this.idFun(dstNode);
        const srcIdx = this.id2idx.get(srcID);
        const dstIdx = this.id2idx.get(dstID);

        if (this.links.some(lk => lk.source === srcIdx && lk.target === dstIdx))
            return;

        this.links.push({source: srcIdx, target: dstIdx});
        this.len.set([srcID, dstID].join("~"), 1);

        for (const [key, value] of this.len.entries()) {
            const [sid, did] = key.split('~');
            if (did === srcID)
                this.len.set([sid, dstID].join("~"), value + 1);
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
}

console.info("DAG.js is loaded");