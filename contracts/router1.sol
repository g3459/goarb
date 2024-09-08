contract Router{

    struct Route{
        uint amOut;
        bytes calls;
    }

    struct TokenInfo{
        uint ethPX64;
        address token;
    }
    
    function findRoutes(uint t,uint amIn,uint[][][] calldata pools) public view returns (Route[] memory routes){
        unchecked{
            routes=new Route[](pools.length);
            routes[t].amOut=amIn;
            findRoutes(pools,routes,2);
        }
    }

    function findRoutes(uint[][][] calldata pools,Route[] memory routes,uint depth) internal view{
        unchecked{
            uint updated=type(uint).max>>(256-routes.length);
            while (updated!=0){
                for (uint t0; t0 < pools.length; t0++){
                    //Si se ha actualizado tIn vuelve a comprovar tIn con todos los tokens
                    if(updated&(1<<t0)!=0){
                        updated^=(1<<t0);
                        //La amIn es la amOut para el tIn
                        uint amIn=routes[t0].amOut;
                        uint ethIn=routes[0].amOut;
                        if(amIn>0){
                            for (uint t1; t1 < pools.length; t1++){
                                if(t0!=t1){
                                    bool direc;
                                    uint[] memory _pools;
                                    if(pools[t0][t1].length>0){
                                        direc=true;
                                        _pools=pools[t0][t1];
                                    }else if(pools[t1][t0].length>0){
                                        _pools=pools[t1][t0];
                                    }else{
                                        continue;
                                    }
                                    Route[][] memory step=new Route[][](depth);
                                    for(uint i; i< depth;i++){
                                        step[i]=new Route[](i+1);
                                    }
                                    //Recorre todas las pools para un mismo par en busca de una mayor amOut para el tOut
                                    uint p;
                                    while(p<_pools.length){
                                        uint rIn;uint rOut;
                                        {
                                            uint slot0=_pools[p++];
                                            rIn=slot0>>128;
                                            rOut=uint128(slot0);
                                        }
                                        uint slot1=_pools[p++];
                                        (rIn,rOut)=updateReserves(routes[t0].calls,rIn, rOut, slot1);
                                        uint rInLimit;
                                        {
                                            uint slot2=_pools[p++];
                                            if(direc){
                                                rInLimit=(slot2>>128);
                                            }else{
                                                rInLimit=uint128(slot2);
                                                (rIn,rOut)=(rOut,rIn);
                                            }
                                        }
                                        for(uint i; i< depth;i++){
                                            uint _amIn=amIn/(i+1);
                                            uint poolCall;
                                            {
                                                uint rshAmIn=_amIn;
                                                {
                                                    uint rsh;
                                                    while(uint48(rshAmIn)!=rshAmIn){
                                                        rsh+=8;
                                                        rshAmIn>>=8;
                                                    }
                                                    rshAmIn<<=8;
                                                    rshAmIn|=rsh;
                                                }
                                                poolCall=(slot1&0x7fffffffff00000000000000ffffffffffffffffffffffffffffffffffffffff)|(rshAmIn<<160);
                                            }
                                            if(direc) poolCall|=0x8000000000000000000000000000000000000000000000000000000000000000;
                                            if(rInLimit!=0&&_amIn+rIn>rInLimit){
                                                continue;
                                            }
                                            uint amOut=_amIn*uint24(slot1>>160);
                                            amOut = (amOut * rOut) / (rIn * 1e6 + amOut);
                                            {
                                                uint gasFee=((uint8(slot1>>216)<2?100000:300000)*tx.gasprice);
                                                if(t1!=0){
                                                    gasFee=(amOut*gasFee)/(ethIn/(i+1));
                                                }
                                                amOut-=gasFee;
                                            }
                                            for(uint j; j <= i;j++){
                                                if(int(amOut)>int(step[i][j].amOut)){
                                                    for(uint k=i;k>j;k--){
                                                        step[i][k]=step[i][k-1];
                                                    }
                                                    step[i][j].amOut=amOut;
                                                    step[i][j].calls=abi.encode(poolCall);
                                                    break;
                                                }
                                            }
                                        }
                                    }
                                    for(uint i;i<depth;i++){
                                        Route memory combined;
                                        for(uint j;j<=i;j++){
                                            combined.amOut+=step[i][j].amOut;
                                            combined.calls=bytes.concat(combined.calls,step[i][j].calls);
                                        }
                                        if(combined.amOut>routes[t1].amOut){
                                            //Actualiza amOut para tOut y Copia tIn calls a tOut calls añadiendo la nueva call
                                            routes[t1].amOut=combined.amOut;
                                            routes[t1].calls=bytes.concat(routes[t0].calls,combined.calls);
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    function concat(bytes memory src, uint slot) public pure returns (bytes memory dst) {
        assembly {
            let lengthBytesMemory := mload(src)
            let totalLength := add(lengthBytesMemory, 0x20)
            dst := mload(0x40)
            mstore(dst, totalLength)
            let dstPtr := add(dst, 0x20)
            let srcPtr := add(src, 0x20)
            for { let i := 0 } lt(i, lengthBytesMemory) { i := add(i, 0x20) } {
                mstore(add(dstPtr, i), mload(add(srcPtr, i)))
            }
            mstore(add(dstPtr, lengthBytesMemory), slot)
            mstore(0x40, add(dst, add(totalLength, 0x20)))
        }
    }

    function updateReserves(bytes memory calls,uint r0,uint r1,uint slot1)internal pure returns(uint,uint){
        uint fees=uint24(slot1>>160);
        uint160 pool=uint160(slot1);
        for (uint i=0x20; i <= calls.length; i += 0x20) {
            uint _poolCall;
            assembly{_poolCall:= mload(add(calls, i))}
            if (pool == uint160(_poolCall)){
                uint amIn=uint(uint48(_poolCall>>168))<<uint8(_poolCall>>160);
                uint amOut=amIn*fees;
                bool v2fee = uint8(slot1>>216)==1;
                if(_poolCall&0x8000000000000000000000000000000000000000000000000000000000000000==0){
                    r0-=(amOut*r0)/(r1*1e6+amOut);
                    r1+=v2fee?amIn:(amOut/1e6);
                }else{
                    r1-=(amOut*r1)/(r0*1e6+amOut);
                    r0+=v2fee?amIn:(amOut/1e6);
                }
            }
        }
        return (r0,r1);
    }
}
