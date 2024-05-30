contract Arouter{

    struct Route{
        uint amOut;
        bytes calls;
    }

    struct Routes{
        uint amIn;
        Route[] routes;
    }
    
    struct UniProtocol{
        uint24[] fees;
        address factory;
        bytes32 poolInitCode;
    }

    // function findPoolsTest()public view returns(bytes[][] memory pools){
    //     address[] memory tokens=new address[](2);
    //     tokens[0]=0xc2132D05D31c914a87C6611C10748AEb04B58e8F;
    //     tokens[1]=0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174;
    //     UniProtocol[] memory protocols=new UniProtocol[](1);
    //     protocols[0].fees=new uint24[](1);
    //     protocols[0].fees[0]=uint24(100);
    //     protocols[0].factory=0x1F98431c8aD98523631AE4a59f267346ea31F984;
    //     protocols[0].poolInitCode=bytes32(0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54);
    //     return this.findPools(tokens,protocols);
    // }


    function findPools(address[] calldata tokens, UniProtocol[] calldata protocols)public view returns(bytes[][] memory pools){
        unchecked {
            pools=new bytes[][](tokens.length);
            for (uint t0; t0 < tokens.length; t0++)
                pools[t0]=new bytes[](tokens.length);
            bytes32 fmp;
            assembly{
                fmp:=mload(0x40)
            }
            for (uint t0; t0 < tokens.length; t0++){
                for (uint t1; t1 < tokens.length; t1++){
                    if(tokens[t0]<tokens[t1]){
                        bytes32 stp;
                        assembly{
                            stp:=fmp
                            fmp:=add(fmp,0x20)
                        }
                        for (uint p; p < protocols.length; p++) {
                            for (uint f; f < protocols[p].fees.length; f++) {
                                UniProtocol memory protocol=protocols[p];
                                address pool=address(uint160(uint(keccak256(abi.encodePacked(hex'ff',protocol.factory, keccak256(abi.encode(tokens[t0], tokens[t1],protocol.fees[f])),protocol.poolInitCode)))));
                                if (pool.code.length > 0) {
                                    uint liquidity=IUniV3Pool(pool).liquidity();
                                    if(liquidity>2){
                                        uint reserve0;uint reserve1;uint reserve0Limit;uint reserve1Limit;bytes32 stateHash;
                                        {
                                            (, bytes memory state) = pool.staticcall(abi.encodeWithSelector(IUniV3Pool.slot0.selector));
                                            stateHash=keccak256(state);
                                            int t;
                                            int s = IUniV3Pool(pool).tickSpacing();
                                            {
                                                uint sqrtPX64;
                                                (sqrtPX64,t) = abi.decode(state, (uint, int));
                                                sqrtPX64>>=32;
                                                reserve0=(liquidity<<64)/sqrtPX64;
                                                reserve1=(liquidity*sqrtPX64)>>64;
                                            }
                                            reserve0Limit=((liquidity<<64)/tSqrtPX64(t < 0 ? int((t + 1) / s - 1) * s : int(t / s) * s)) - reserve0;
                                            reserve1Limit=((liquidity*tSqrtPX64(t < 0 ? int((t + 1) / s) * s : int(t / s + 1) * s))>>64) - reserve1;
                                        }
                                        if(reserve0>reserve0Limit&&reserve1>reserve1Limit){
                                            uint fee=1e6-protocol.fees[f];
                                            assembly{
                                                mstore(fmp,or(shl(128,reserve0),reserve1))
                                                fmp:=add(fmp,0x20)
                                                mstore(fmp,or(shl(128,reserve0Limit),reserve1Limit))
                                                fmp:=add(fmp,0x20)
                                                mstore(fmp,or(or(and(stateHash,0xfffffffe00000000000000000000000000000000000000000000000000000000),shl(160,fee)),and(pool,0x000000000000000000000000ffffffffffffffffffffffffffffffffffffffff)))
                                                fmp:=add(fmp,0x20)
                                            }
                                        }
                                    }
                                }
                            }
                        }
                        bytes memory _pools;
                        assembly{
                            _pools:=stp
                            mstore(_pools,sub(sub(fmp,stp),0x20))
                        }
                        pools[t1][t0]=pools[t0][t1]=_pools;
                    }
                }
            }
            assembly{mstore(0x40,fmp)}
        }
    }

    function allTokensWithBalances(address[] calldata tokens,UniProtocol[] calldata protocols,uint[] calldata ethPricesX64,uint minEth) public view returns (Routes[][] memory routes){
        unchecked{
            routes=new Routes[][](tokens.length);
            bytes[][] memory pools=findPools(tokens,protocols);
            for(uint i;i<tokens.length;i++){
                uint b = IERC20(tokens[i]).balanceOf(msg.sender);
                uint n;
                {
                    uint _b=b;
                    uint minEthX64=uint(minEth)<<64;
                    while(_b*ethPricesX64[i]>minEthX64){
                        n++;
                        _b>>=1;
                    }
                }
                routes[i]=new Routes[](n);
                for(uint k; k<n;k++){
                    routes[i][k].amIn=b;
                    routes[i][k].routes=new Route[](tokens.length);
                    uint gasPQ=((b*ethPricesX64[i]) / (tx.gasprice>0?tx.gasprice:(block.basefee+30e9)))>>64;
                    routes[i][k].routes[i].amOut=b-(b*21000)/gasPQ;
                    findRoutes(tokens,routes[i][k].routes,/*k>0?routes[i][k-1].routes:new Route[](0),*/pools,gasPQ);
                    b>>=1;
                }
            }
        }
    }

    function singleToken(address[] calldata tokens,UniProtocol[] calldata protocols,uint ethPriceInX64,uint amIn,uint tIn) public view returns (Route[] memory routes){
        unchecked{
            routes=new Route[](tokens.length);
            uint gasPQ=((amIn*ethPriceInX64) / (tx.gasprice>0?tx.gasprice:(block.basefee+30e9)))>>64;
            routes[tIn].amOut=amIn-(amIn*21000)/gasPQ;
            bytes[][] memory pools=findPools(tokens,protocols);
            findRoutes(tokens,routes,pools,gasPQ);
        }
    }

    function findRoutes(address[] calldata tokens,Route[] memory routes,bytes[][] memory pools,uint gasPQ) internal pure{
        unchecked{
            bytes32 updated;
            while(true){
                for (uint t0; t0 < tokens.length; t0++){
                    updated|=bytes32(0x01<<t0);
                    if(routes[t0].amOut>0){
                        Route memory routeIn = routes[t0];
                        for (uint t1; t1 < tokens.length; t1++){
                            if(t0!=t1){
                                Route memory routeOut = routes[t1];
                                bool direc = tokens[t0] < tokens[t1];
                                bytes memory _pools=pools[t0][t1];
                                for(uint p;p<_pools.length;p+=0x60){
                                    uint amIn=routeIn.amOut;
                                    uint slot0;uint slot2;
                                    assembly{
                                        slot0:=mload(add(add(_pools,0x20),p))
                                        slot2:=mload(add(add(_pools,0x60),p))
                                    }
                                    if((direc?(slot0>>128):uint128(slot0))>amIn<<3 && !checkPool(routeIn.calls,address(uint160(slot2)))){
                                        amIn-=(amIn*85000)/gasPQ;
                                        uint amOut=amIn*uint24(slot2>>160);
                                        amOut = (direc
                                            ? (amOut * uint128(slot0)) / ((slot0>>128) * 1e6 + amOut)
                                            : (amOut * (slot0>>128)) / (uint128(slot0) * 1e6 + amOut));
                                        uint slot1;
                                        assembly{
                                            slot1:=mload(add(add(_pools,0x40),p))
                                        }
                                        if((direc?uint128(slot1):(slot1>>128))>amOut){
                                            if(amOut>routeOut.amOut){
                                                routeOut.amOut=amOut;
                                                bytes memory rInCalls=routeIn.calls;
                                                bytes memory rOutCalls=routeOut.calls;
                                                uint rLen=0;
                                                while(rLen<rInCalls.length){
                                                    bytes32 _call;
                                                    assembly {
                                                        _call := mload(add(rInCalls, rLen))
                                                    }
                                                    if(_call==0){
                                                        break;
                                                    }
                                                    rLen+=0x20;
                                                }
                                                rLen+=0x20;
                                                
                                                if(rLen>rOutCalls.length){
                                                    rOutCalls=(routeOut.calls=new bytes(rLen));
                                                }else{
                                                    for(uint i=rLen;i<rOutCalls.length;i+=0x20){
                                                        assembly{
                                                            mstore(add(add(rOutCalls,0x20),i),0)
                                                        }
                                                    }
                                                }
                                                for(uint i=0x20;i<=rInCalls.length;i+=0x20){
                                                    assembly{
                                                        mstore(add(rOutCalls,i),mload(add(rInCalls,i)))
                                                    }
                                                }
                                                uint callbytes=slot2&0xffffffff0000000000000000ffffffffffffffffffffffffffffffffffffffff;
                                                if(direc) callbytes|=0x0000000001000000000000000000000000000000000000000000000000000000;
                                                assembly{
                                                    mstore(add(rOutCalls,rLen),callbytes)
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
                if(updated==bytes32(uint(0x01<<tokens.length)-1)) return;
            }
        }
    }

    function checkPool(bytes memory calls, address pool) internal pure returns (bool) {
        unchecked {
            for (uint i; i < calls.length; i += 0x20) {
                address _pool;
                assembly {
                    _pool := mload(add(add(calls, 0x20), i))
                }
                if (pool == _pool) return true;
            }
            return false;
        }
    }

    // function test()public view returns(bool zz,address aaaa,uint slot,uint128 aa,bytes32 b,address pool,uint24 fee,bytes4 stateHash){
    //     b|=bytes4(0x00000001);
    //     // b=bytes32(type(uint).max&type(uint128).max);
    //     // slot=uint(0x000000000000000000000000000000000000000000000006815e0ff03491b2770255d);
    //     // aa=uint128(slot>>128);
    //     // bytes32 fmp;
    //     // bytes4 _stateHash=0x29865dc8;
    //     // address _pool=0xDaC8A8E6DBf8c690ec6815e0fF03491B2770255D;
    //     // uint _fee=1e6-100;
    //     // assembly{
    //     //     aaaa:=slot
    //     //     fmp:=mload(0x40)
    //     //     mstore(fmp,0x2713753a00000000000f41dcdac8a8e6dbf8c690ec6815e0ff03491b2770255d)
    //     //     pool:=mload(fmp)
    //     //     fee:=shr(160,pool)
    //     //     stateHash:=pool
    //     // }
    // }

    function tSqrtPX64(int t) internal pure returns(uint) {
        unchecked{
            uint abst = t < 0 ? uint(-t) : uint(t);
            uint ratio = abst & 0x1 != 0 ? 0xfffcb933bd6fad37aa2d162d1a594001 : 0x100000000000000000000000000000000;
            if (abst & 0x2 != 0) ratio = (ratio * 0xfff97272373d413259a46990580e213a) >> 128;
            if (abst & 0x4 != 0) ratio = (ratio * 0xfff2e50f5f656932ef12357cf3c7fdcc) >> 128;
            if (abst & 0x8 != 0) ratio = (ratio * 0xffe5caca7e10e4e61c3624eaa0941cd0) >> 128;
            if (abst & 0x10 != 0) ratio = (ratio * 0xffcb9843d60f6159c9db58835c926644) >> 128;
            if (abst & 0x20 != 0) ratio = (ratio * 0xff973b41fa98c081472e6896dfb254c0) >> 128;
            if (abst & 0x40 != 0) ratio = (ratio * 0xff2ea16466c96a3843ec78b326b52861) >> 128;
            if (abst & 0x80 != 0) ratio = (ratio * 0xfe5dee046a99a2a811c461f1969c3053) >> 128;
            if (abst & 0x100 != 0) ratio = (ratio * 0xfcbe86c7900a88aedcffc83b479aa3a4) >> 128;
            if (abst & 0x200 != 0) ratio = (ratio * 0xf987a7253ac413176f2b074cf7815e54) >> 128;
            if (abst & 0x400 != 0) ratio = (ratio * 0xf3392b0822b70005940c7a398e4b70f3) >> 128;
            if (abst & 0x800 != 0) ratio = (ratio * 0xe7159475a2c29b7443b29c7fa6e889d9) >> 128;
            if (abst & 0x1000 != 0) ratio = (ratio * 0xd097f3bdfd2022b8845ad8f792aa5825) >> 128;
            if (abst & 0x2000 != 0) ratio = (ratio * 0xa9f746462d870fdf8a65dc1f90e061e5) >> 128;
            if (abst & 0x4000 != 0) ratio = (ratio * 0x70d869a156d2a1b890bb3df62baf32f7) >> 128;
            if (abst & 0x8000 != 0) ratio = (ratio * 0x31be135f97d08fd981231505542fcfa6) >> 128;
            if (abst & 0x10000 != 0) ratio = (ratio * 0x9aa508b5b7a84e1c677de54f3e99bc9) >> 128;
            if (abst & 0x20000 != 0) ratio = (ratio * 0x5d6af8dedb81196699c329225ee604) >> 128;
            if (abst & 0x40000 != 0) ratio = (ratio * 0x2216e584f5fa1ea926041bedfe98) >> 128;
            if (abst & 0x80000 != 0) ratio = (ratio * 0x48a170391f7dc42444e8fa2) >> 128;
            if (t > 0) ratio = type(uint).max / ratio;
            return ratio >> 64;
        }
    }
}



//interfaces

interface IUniV2Pool{
    function swap(uint amount0Out, uint amount1Out, address to, bytes calldata data) external;
    function getReserves()external view returns(uint reserve0, uint reserve1, uint blockTimestampLast);
}

interface IUniV3Pool{
    function swap(address recipient, bool zeroForOne, int amountSpecified, uint160 sqrtPriceLimitX96, bytes calldata data) external returns(int amount0, int amount1);
    function fee() external view returns(uint24 fee);
    function slot0() external view returns(uint sqrtPX96, int t, uint observationIndex, uint observationCardinality, uint observationCardinalityNext, uint feeProtocol, bool unlocked);
    function liquidity() external view returns(uint liquidity);
    function tickSpacing() external view returns(int s);
    function flash(address recipient, uint amount0, uint amount1, bytes calldata data) external;
}

interface IAlgebraV3Pool{
    function swap(address recipient, bool zeroForOne, int amountSpecified, uint160 sqrtPriceLimitX96, bytes calldata data) external returns(int amount0, int amount1);
    function globalState() external view returns(uint sqrtPX96, int t, uint fee, uint timePointIndex, uint comunityFeet0, uint comunityFeeT1, bool unlocked);
    function tickSpacing() external view returns(int s);
    function liquidity() external view returns(uint liquidity);
}

interface IERC20{
    function balanceOf(address ) external view returns ( uint256 );
    function transfer(address, uint256 ) external returns ( uint256 );
}


