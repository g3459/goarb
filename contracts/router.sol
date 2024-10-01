library CRouter{

    function findRoutesInt(uint8 maxLen,uint8 t,uint amIn,bytes[][] memory pools) internal view returns (uint[] memory amounts,bytes[] memory calls){
        unchecked{
            return findRoutes(maxLen, t, amIn, pools);
        }
    }

    function findRoutes(uint8 maxLen,uint8 t,uint amIn,bytes[][] memory pools) public view returns (uint[] memory amounts,bytes[] memory calls){
        unchecked{
            amounts=new uint[](pools.length);
            amounts[t]=amIn;
            calls=new bytes[](pools.length);
            findRoutes(maxLen*0x20,pools,amounts,calls);
        }
    }

    function findRoutes(uint8 maxLen,bytes[][] memory pools,uint[] memory amounts,bytes[] memory calls) internal view{
        unchecked{
            uint updated=type(uint).max>>(256-calls.length);
            while (updated!=0){
                for (uint t0; t0 < pools.length; t0++){
                    //Si se ha actualizado tIn vuelve a comprovar tIn con todos los tokens
                    if(updated&(1<<t0)==0){
                        continue;
                    }
                    updated^=1<<t0;
                    //La amounts[t0] es la amOut para el tIn
                    if(amounts[t0]==0 || calls[t0].length==maxLen){
                        continue;
                    }
                    for (uint t1; t1 < pools.length; t1++){
                        if(t0==t1){
                            continue;
                        }
                        bool direc;
                        bytes memory _pools;
                        if(pools[t0].length>0 && pools[t0][t1].length>0){
                            direc=true;
                            _pools=pools[t0][t1];
                        }else if(pools[t1].length>0 && pools[t1][t0].length>0){
                            _pools=pools[t1][t0];
                        }else{
                            continue;
                        }
                        uint poolCall;
                        //Recorre todas las pools para un mismo par en busca de una mayor amOut para el tOut
                        uint p;
                        while(p<_pools.length){
                            uint rIn;uint rOut;
                            {
                                p+=0x20;
                                uint slot0;//=uint(bytes32(_pools[p:]));
                                assembly{
                                    slot0:=mload(add(_pools,p))
                                }
                                rIn=slot0>>128;
                                rOut=uint128(slot0);
                            }
                            // p+=0x20;
                            if(!direc){
                                (rIn,rOut)=(rOut,rIn);
                            }
                            p+=0x20;
                            uint slot1;//=uint(bytes32(_pools[p:]));
                            assembly{
                                slot1:=mload(add(_pools,p))
                            }
                            //p+=0x20;
                            {
                                p+=0x20;
                                uint slot2;//=uint(bytes32(_pools[p:]));
                                assembly{
                                    slot2:=mload(add(_pools,p))
                                }
                                // p+=0x20;
                                if(slot2!=0 && amounts[t0]+rIn>(direc?(slot2>>128):uint128(slot2))){
                                    continue;
                                }
                            }
                            if(poolInCalls(calls[t0],uint160(slot1))){
                                continue;
                            }
                            uint amInXFee= amounts[t0] * uint24(slot1>>160);
                            uint amOut = (amInXFee * rOut) / (rIn * 1e6 + amInXFee);
                            {
                                uint gasFee=(uint8(slot1>>216)==2?300000:100000)*tx.gasprice;
                                if (t1!=0){
                                    gasFee=(amOut*gasFee)/amounts[0];
                                }
                                amOut-=gasFee;
                                if(int(amOut)<=int(amounts[t1])){
                                    continue;
                                }
                                uint amOutX2 = amInXFee<<1;
                                amOutX2 = (amOutX2 * rOut) / (rIn * 1e6 + amOutX2);
                                if(int(amOutX2-gasFee)>int(amOut<<1)){
                                    continue;
                                }
                            }
                            // hAmOut=amOut;
                            amounts[t1]=amOut;
                            poolCall=slot1;
                        }
                        if(poolCall==0){
                            continue;
                        }
                        //Actualiza amOut para tOut y Copia tIn calls a tOut calls añadiendo la nueva call
                        //amounts[t1]=hAmOut;
                        uint amIn56bit=compress56bit(amounts[t0]);
                        poolCall=(poolCall&0x7fffffffff00000000000000ffffffffffffffffffffffffffffffffffffffff)|(amIn56bit<<160);
                        if(direc) poolCall|=0x8000000000000000000000000000000000000000000000000000000000000000;
                        calls[t1]=bytes.concat(calls[t0],abi.encode(poolCall));
                        updated|=1<<t1;
                    }
                }
            }
        }
    }

    // function decompress56bit(uint compressed)internal pure returns (uint){
    //     unchecked{
    //         return uint(uint48(compressed>>8))<<uint8(compressed);
    //     }
    // }

    function compress56bit(uint uncompressed)internal pure returns(uint){
        unchecked{
            uint temp=uncompressed;
            uint rsh;
            while(uint48(temp)!=temp){
                rsh+=8;
                temp>>=8;
            }
            temp<<=8;
            temp|=rsh;
            return temp;
        }
    }

    function poolInCalls(bytes memory calls,uint160 pool)internal pure returns(bool){
        // uint fees=uint24(slot1>>160);
        // uint160 pool=uint160(slot1);
        for (uint i=0x20; i <= calls.length; i += 0x20) {
            uint _poolCall;
            assembly{_poolCall:= mload(add(calls, i))}
            if (pool == uint160(_poolCall)){
                return true;
                // uint amounts[t0]=decompress56bit(_poolCall>>160);
                // uint amounts[t0]XFee=amounts[t0]*fees;
                // bool v2fee = uint8(slot1>>216)==1;
                // if(_poolCall&0x8000000000000000000000000000000000000000000000000000000000000000==0){
                //     r0-=(amounts[t0]XFee*r0)/(r1*1e6+amounts[t0]XFee);
                //     r1+=v2fee?amounts[t0]:(amounts[t0]XFee/1e6);
                // }else{
                //     r1-=(amounts[t0]XFee*r1)/(r0*1e6+amounts[t0]XFee);
                //     r0+=v2fee?amounts[t0]:(amounts[t0]XFee/1e6);
                // }
            }
        }
        return false;//(r0,r1);
    }
}
