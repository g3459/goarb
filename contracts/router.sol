contract Router{

    struct Route{
        uint amOut;
        bytes calls;
    }

    struct TokenInfo{
        uint ethPX64;
        address token;
    }
    
    function findRoutes(TokenInfo[] calldata tokens,uint[][][] calldata pools,uint depth,uint amIn,uint t) public view returns (Route[] memory routes){
        unchecked{
            routes=new Route[](tokens.length);
            routes[t].amOut=amIn;
            findRoutes(tokens,pools,depth,routes);
        }
    }

    function findRoutes(TokenInfo[] calldata tokens,uint[][][] calldata pools,uint depth,Route[] memory routes) internal view{
        unchecked{
            uint updated=type(uint).max>>(256-tokens.length);
            while(true){
                for (uint t0; t0 < tokens.length; t0++){
                    //Si se ha actualizado tIn vuelve a comprovar tIn con todos los tokens
                    if(updated&(1<<t0)!=0){
                        updated^=(1<<t0);
                        //La amIn es la amOut para el tIn
                        if(routes[t0].amOut>0){
                            {
                                //Busca 2 subrutas para la mitad de la amIn.
                                uint _amIn=routes[t0].amOut>>1;
                                if(depth>0){
                                    Route[] memory subRoutes1=new Route[](tokens.length);
                                    subRoutes1[t0]=Route(_amIn,routes[t0].calls);
                                    findRoutes(tokens,pools,depth-1,subRoutes1);
                                    Route[] memory subRoutes2=new Route[](tokens.length);
                                    for(uint i;i<tokens.length;i++)
                                        subRoutes2[i].calls=subRoutes1[i].calls;
                                    subRoutes2[t0].amOut=_amIn;
                                    findRoutes(tokens,pools,depth-1,subRoutes2);
                                    for (uint t1; t1 < tokens.length; t1++){
                                        //Si la combinacion de estas dos subrutas es mejor que la actual ruta se substituye.
                                        uint amOut=subRoutes1[t1].amOut+subRoutes2[t1].amOut;
                                        if(amOut>routes[t1].amOut){
                                            routes[t1]=Route(amOut,subRoutes2[t1].calls);
                                            assembly{updated:=or(updated,shl(t1,0x01))}
                                        }
                                    }
                                }
                            }
                            uint amIn=routes[t0].amOut;
                            for (uint t1; t1 < tokens.length; t1++){
                                if(t0!=t1){
                                    uint[] memory _pools=pools[t0][t1];
                                    if(pools.length>0){
                                        uint poolCall;
                                        uint ethPX64=tokens[t1].ethPX64;
                                        uint hAmOut=routes[t1].amOut;
                                        bool direc = tokens[t0].token < tokens[t1].token;
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
                                            {
                                                bytes memory calls=routes[t0].calls;
                                                for (uint i=0x20; i < calls.length; i += 0x20) {
                                                    uint _poolCall;
                                                    assembly{_poolCall:= mload(add(calls, i))}
                                                    if (uint160(poolCall) == uint160(_poolCall)){
                                                        slot1&=0xffffffffffffffffffffffffffffffffffffffffffffffffffffffff;
                                                        uint _amIn=uint(uint48(_poolCall>>168))<<uint8(_poolCall>>160);
                                                        uint _amOut=_amIn*uint24(slot1>>160);
                                                        assembly{
                                                            switch and(_poolCall,0x8000000000000000000000000000000000000000000000000000000000000000)
                                                            case 0{
                                                                rOut:= add(rOut,_amIn)
                                                                rIn:= sub(rIn,div(mul(_amOut,rIn),add(mul(rOut,1000000),_amOut)))
                                                            }default{
                                                                rIn:= add(rIn,_amIn)
                                                                rOut:= sub(rOut,div(mul(_amOut,rOut),add(mul(rIn,1000000),_amOut)))
                                                            }
                                                        }
                                                    }
                                                }
                                            }
                                            if (!direc) (rIn,rOut)=(rOut,rIn);
                                            uint8 t=uint8(slot1>>216);
                                            if(t!=1){
                                                uint slot2=_pools[p++];
                                                if(rIn+amIn>(direc?(slot2>>128):uint128(slot2)))
                                                    continue;
                                            }
                                            uint amOut=amIn*uint24(slot1>>160);
                                            amOut = (amOut * rOut) / (rIn * 1e6 + amOut);
                                            if(amOut>hAmOut){
                                                amOut-=((t<2?(90000<<64):(285000<<64))*tx.gasprice)/ethPX64;
                                                if (int(amOut)>int(hAmOut)){
                                                    (hAmOut,poolCall)=(amOut,slot1);
                                                }
                                            }
                                            
                                        }
                                        if(poolCall!=0){
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
                if(updated==0) return;
            }
        }
    }
}