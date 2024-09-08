contract Router{

    struct Route{
        uint amOut;
        bytes calls;
    }

    struct TokenInfo{
        uint ethPX64;
        address token;
    }
    
    function findRoutes(uint maxLen,uint t,uint amIn,uint[][][] calldata pools) public view returns (Route[] memory routes){
        unchecked{
            routes=new Route[](pools.length);
            routes[t].amOut=amIn;
            findRoutes(maxLen,pools,routes);
        }
    }

    function findRoutes(uint maxLen,uint[][][] calldata pools,Route[] memory routes) internal view{
        unchecked{
            uint updated=type(uint).max>>(256-routes.length);
            while (updated!=0){
                for (uint t0; t0 < pools.length; t0++){
                    //Si se ha actualizado tIn vuelve a comprovar tIn con todos los tokens
                    if(updated&(1<<t0)!=0){
                        updated^=(1<<t0);
                        //La amIn es la amOut para el tIn
                        uint amIn=routes[t0].amOut;
                        if(amIn>0 && routes[t0].calls.length<maxLen){
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
                                    uint poolCall;
                                    uint hAmOut=routes[t1].amOut;
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
                                        if(rInLimit==0||amIn+rIn<rInLimit){
                                            uint amOut = (amIn * rOut) / (rIn + amIn);
                                            if(amOut>hAmOut){
                                                amOut=(amOut*uint24(slot1>>160))/1e6;
                                                {
                                                    uint gasFee=((uint8(poolCall>>216)<2?100000:300000)*tx.gasprice);
                                                    if (t1!=0){
                                                        gasFee=(amOut*gasFee)/routes[0].amOut;
                                                    }
                                                    amOut-=gasFee;
                                                }
                                                if(int(amOut)>int(hAmOut)){
                                                    hAmOut=amOut;
                                                    poolCall=slot1;
                                                }
                                            }
                                        }
                                    }
                                    if(poolCall!=0){
                                        //Actualiza amOut para tOut y Copia tIn calls a tOut calls añadiendo la nueva call
                                        routes[t1].amOut=hAmOut;
                                        uint amIn48bit=amIn;
                                        {
                                            uint rsh;
                                            while(uint48(amIn48bit)!=amIn48bit){
                                                rsh+=8;
                                                amIn48bit>>=8;
                                            }
                                            amIn48bit<<=8;
                                            amIn48bit|=rsh;
                                        }
                                        poolCall=(poolCall&0x7fffffffff00000000000000ffffffffffffffffffffffffffffffffffffffff)|(amIn48bit<<160);
                                        if(direc) poolCall|=0x8000000000000000000000000000000000000000000000000000000000000000;
                                        routes[t1].calls=bytes.concat(routes[t0].calls,abi.encode(poolCall));
                                        assembly{updated:=or(updated,shl(t1,0x01))}
                                    }
                                }
                            }
                        }
                    }
                }
            }
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
