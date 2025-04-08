object "Caller" {
    code {
		// constructor
        let stp:=datasize("runtime")
		datacopy(0, dataoffset("runtime"), stp)
        
        // mstore(stp,0x1F98431c8aD98523631AE4a59f267346ea31F984)
        let protLen:=2
        mstore(add(stp,mul(0x14,add(15,protLen))),0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2)
        mstore(add(stp,mul(0x14,add(14,protLen))),0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48)
        mstore(add(stp,mul(0x14,add(13,protLen))),0xdAC17F958D2ee523a2206206994597C13D831ec7)

        mstore(add(stp,mul(0x14,1)),0x1F98431c8aD98523631AE4a59f267346ea31F984)
        mstore(add(stp,mul(0x14,0)),0x1F98431c8aD98523631AE4a59f267346ea31F984)
        
		return(0, add(stp,mul(add(16,protLen),0x14)))
	}
    object "runtime"{
        code {
            let callData := calldataload(0)
            let amIn
            amIn:=shl(shr(0xf8,callData),1)
            codecopy(0x00, sub(codesize(),mul(add(shr(0xfc,callData),1),0x14)),0x14)
            callData := shl(4,callData)
            let t0:=mload(0x00)
            codecopy(0x00, sub(codesize(),mul(add(shr(0xfc,callData),1),0x14)),0x14)
            callData := shl(4,callData)
            let t1:=mload(0x00)
            let pid:=shr(0xf8,callData)
            codecopy(0x00, sub(codesize(),mul(add(shr(4, pid),17),0x14)),0x14)
            callData := shl(8,callData)
            let factory:=mload(0x00)
            pid:=and(0x0f,pid)
            if lt(pid,5){
                let f
                switch pid
                case 1{
                    f:=100
                }
                case 2{
                    f:=500
                }
                case 3{
                    f:=3000
                }
                case 4{
                    f:=10000
                }
                mstore(0x00, 0x1698ee82)
                mstore(0x04, t0)
                mstore(0x24, t1)
                mstore(0x44, f)
                pop(staticcall(gas(), factory, 0x00, 0x64, 0x00, 0x20))
                let pool := mload(0x00)
                mstore(0x00, 0x128acb08)
                mstore(0x04, address())
                mstore(0x24, lt(t0,t1))
                mstore(0x44, amIn)
                let sqrtL:=4295128740
                if lt(t1,t0){
                    sqrtL:=1461446703485210103287273052203988822378723970341
                }
                mstore(0x64, sqrtL)
                mstore(0x84, 0xa0)
                if iszero(call(gas(), pool, 0, 0, 0xc4, 0, 0x40)) {
                    revert(0, 0)
                }
                return(0,0)
            }
        }
    }
}
