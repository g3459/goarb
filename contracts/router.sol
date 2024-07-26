contract Router{

    struct Route{
        uint amOut;
        bytes calls;
    }

    struct TokenInfo{
        uint ethPX64;
        address token;
    }
    
    function findRoutes(uint depth,uint amIn,uint t,TokenInfo[] calldata tokens,uint[][][] calldata pools) public view returns (Route[] memory routes){
        unchecked{
            routes=new Route[](tokens.length);
            routes[t].amOut=amIn;
            bytes[] memory upCalls;
            findRoutes(depth,tokens,pools,upCalls,routes);
        }
    }

    function findRoutes(uint depth,TokenInfo[] calldata tokens,uint[][][] calldata pools,bytes[] memory upCalls,Route[] memory routes) internal view{
        unchecked{
            uint updated=type(uint).max>>(256-tokens.length);
            while(true){
                for (uint t0; t0 < tokens.length; t0++){
                    //Si se ha actualizado tIn vuelve a comprovar tIn con todos los tokens
                    if(updated&(1<<t0)!=0){
                        updated^=(1<<t0);
                        //La amIn es la amOut para el tIn
                        uint amIn=routes[t0].amOut;
                        if(amIn>0){
                            if(depth>0){
                                uint _depth=depth-1;
                                uint _amIn=amIn>>1;
                                bytes[] memory _upCalls=new bytes[](tokens.length);
                                Route[] memory _routes1=new Route[](tokens.length);
                                _routes1[t0].amOut=_amIn;
                                for(uint i;i<tokens.length;i++){
                                    _upCalls[i]=routes[i].calls;
                                }
                                findRoutes(_depth,tokens,pools,_upCalls,_routes1);
                                for(uint i;i<tokens.length;i++){
                                    _upCalls[i]=bytes.concat(routes[i].calls,_routes1[i].calls);
                                }
                                Route[] memory _routes2=new Route[](tokens.length);
                                _routes2[t0].amOut=_amIn;
                                findRoutes(_depth,tokens, pools,_upCalls,_routes2);
                                for(uint t1;t1<tokens.length;t1++){
                                    uint amOut=_routes1[t1].amOut+_routes2[t1].amOut;
                                    if(amOut>routes[t1].amOut){
                                        routes[t1].amOut=amOut;
                                        routes[t1].calls=bytes.concat(_upCalls[t1],_routes2[t1].calls);
                                    }
                                }
                            }
                            for (uint t1; t1 < tokens.length; t1++){
                                if(t0!=t1){
                                    bool direc = tokens[t0].token < tokens[t1].token;
                                    uint[] calldata _pools;
                                    if(direc){
                                        if(pools[t0].length==0) continue;
                                        _pools=pools[t0][t1];
                                    }else{
                                        if(pools[t1].length==0) continue;
                                        _pools=pools[t1][t0];
                                    }
                                    if(_pools.length!=0){
                                        uint poolCall;
                                        uint hAmOut;
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
                                            if(upCalls.length!=0)
                                                (rIn,rOut)=updateReserves(upCalls[t1],rIn, rOut, slot1);
                                            (rIn,rOut)=updateReserves(routes[t1].calls,rIn, rOut, slot1);
                                            // if(_rIn!=rIn){
                                            //     poolCall&=0xffffffffffffffffffffffffffffffffffffffffffffffffffffffff;
                                            // }
                                            // (rIn,rOut)=(_rIn,_rOut);
                                            if (!direc) (rIn,rOut)=(rOut,rIn);
                                            if(uint8(slot1>>216)!=1){
                                                uint slot2=_pools[p++];
                                                if(rIn+amIn>(direc?(slot2>>128):uint128(slot2)))
                                                    continue;
                                            }
                                            uint amOut=amIn*uint24(slot1>>160);
                                            amOut = (amOut * rOut) / (rIn * 1e6 + amOut);
                                            if(amOut>hAmOut){
                                                (hAmOut,poolCall)=(amOut,slot1);
                                            }
                                        }
                                        if(hAmOut!=0){
                                            hAmOut-=((uint8(poolCall>>216)<2?(100000<<64):(285000<<64))*tx.gasprice)/tokens[t1].ethPX64;
                                            if(int(hAmOut)>int(routes[t1].amOut)){
                                                //Actualiza amOut para tOut y Copia tInCalls a tOutCalls y le añade la nueva poolCall
                                                routes[t1].amOut=hAmOut;
                                                bytes memory tOutCalls=routes[t1].calls;
                                                uint rLen=routes[t0].calls.length+0x20;
                                                {
                                                    uint len=tOutCalls.length;
                                                    if(len>0){
                                                        while(len<rLen){
                                                            uint nextSlot;
                                                            assembly{nextSlot:=mload(add(add(tOutCalls,0x20),len))}
                                                            if(nextSlot>0) break;
                                                            len+=32;
                                                        }
                                                    }
                                                    if(rLen>len){
                                                        for(uint i;i<tOutCalls.length;i+=32)
                                                            assembly{mstore(add(tOutCalls,i),0)}
                                                        tOutCalls=(routes[t1].calls=new bytes(rLen));
                                                    }else{
                                                        assembly{
                                                            let fm:=mload(0x40)
                                                            let rm:=add(add(tOutCalls,0x20),rLen)
                                                            if gt(rm,fm){mstore(0x40,rm)}
                                                        }
                                                        for(uint i=rLen;i<tOutCalls.length;i+=32)
                                                            assembly{mstore(add(add(tOutCalls,0x20),i),0)}
                                                        if(tOutCalls.length!=rLen) 
                                                            assembly{mstore(tOutCalls,rLen)}
                                                    }
                                                }
                                                {
                                                    bytes memory tInCalls=routes[t0].calls;
                                                    for (uint i=0x20;i<rLen;i+=0x20)
                                                        assembly{mstore(add(tOutCalls,i),mload(add(tInCalls,i)))}
                                                }
                                                uint _amIn=amIn;
                                                {
                                                    uint rsh;
                                                    while(uint48(_amIn)!=_amIn){
                                                        rsh+=8;
                                                        _amIn>>=8;
                                                    }
                                                    _amIn<<=8;
                                                    _amIn|=rsh;
                                                }
                                                poolCall=(poolCall&0x7fffffffff00000000000000ffffffffffffffffffffffffffffffffffffffff)|(_amIn<<160);
                                                if(direc) poolCall|=0x8000000000000000000000000000000000000000000000000000000000000000;
                                                assembly{mstore(add(tOutCalls,rLen),poolCall)}
                                                assembly{updated:=or(updated,shl(t1,0x01))}
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
                if(updated==0) return;
            }
        }
    }

    function updateReserves(bytes memory calls,uint r0,uint r1,uint slot1)internal pure returns(uint,uint){
        uint fees=uint24(slot1>>160);
        uint160 pool=uint160(slot1);
        for (uint i=0x20; i < calls.length; i += 0x20) {
            uint _poolCall;
            assembly{_poolCall:= mload(add(calls, i))}
            if (pool == uint160(_poolCall)){
                uint amIn=uint(uint48(_poolCall>>168))<<uint8(_poolCall>>160);
                uint amOut=amIn*fees;
                bool v2fee = uint8(slot1>>216)==1;
                if(_poolCall&0x8000000000000000000000000000000000000000000000000000000000000000==0){
                    r1+=v2fee?amIn:(amOut/1e6);
                    r0-=(amOut*r0)/(r1*1e6+amOut);
                }else{
                    r0+=v2fee?amIn:(amOut/1e6);
                    r1-=(amOut*r1)/(r0*1e6+amOut);
                }
            }
        }
        return (r0,r1);
    }
}