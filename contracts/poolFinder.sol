contract PoolFinder{

    struct TokenInfo{
        uint ethPX64;
        address token;
    }

    int internal constant MIN_TICK = -887272;
    int internal constant MAX_TICK = 887272;
    bytes32 internal constant B4_MASK = 0xffffffff00000000000000000000000000000000000000000000000000000000;
    bytes32 internal constant ADDR_MASK = 0x000000000000000000000000ffffffffffffffffffffffffffffffffffffffff;

    function findPools(TokenInfo[] calldata tokens,uint minEth)public returns(uint[][][] memory pools){
        unchecked {
            minEth<<=64;
            pools=new uint[][][](tokens.length);
            for (uint t0; t0 < tokens.length; t0++)
                pools[t0]=new uint[][](tokens.length);
            for (uint t0; t0 < tokens.length; t0++){
                address token0=tokens[t0].token;
                uint r0=minEth/tokens[t0].ethPX64;
                for (uint t1; t1 < tokens.length; t1++){
                    address token1=tokens[t1].token;
                    uint r1=minEth/tokens[t1].ethPX64;
                    if(token0<token1){
                        uint[] memory _pools;
                        assembly{
                            _pools:=mload(0x40)
                            mstore(0x40,add(_pools,0x20))
                        }
                        mstoreUniV2Pool(token0,token1,0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32,0x96e8ac4277198ff8b6f785478aa9a39f403cb768dd02cbee326c3e7da348845f,r0,r1);
                        mstoreUniV2Pool(token0,token1,0xc35DADB65012eC5796536bD9864eD8773aBc74C4,0xe18a34eb0e04b04f7a0ac29a6e80748dca96319b42c54d679cb821dca90c6303,r0,r1);
                        mstoreUniV2Pool(token0,token1,0xE7Fb3e833eFE5F9c441105EB65Ef8b261266423B,0xf187ed688403aa4f7acfada758d8d53698753b998a3071b06f1b777f4330eaf3,r0,r1);
                        mstoreUniV3Pool(token0,token1,0x1F98431c8aD98523631AE4a59f267346ea31F984,0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54,r0,r1,100,1);
                        mstoreUniV3Pool(token0,token1,0x1F98431c8aD98523631AE4a59f267346ea31F984,0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54,r0,r1,500,10);
                        mstoreUniV3Pool(token0,token1,0x1F98431c8aD98523631AE4a59f267346ea31F984,0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54,r0,r1,3000,60);
                        mstoreUniV3Pool(token0,token1,0x1F98431c8aD98523631AE4a59f267346ea31F984,0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54,r0,r1,10000,200);
                        mstoreUniV3Pool(token0,token1,0x917933899c6a5F8E37F31E19f92CdBFF7e8FF0e2,0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54,r0,r1,100,1);
                        mstoreUniV3Pool(token0,token1,0x917933899c6a5F8E37F31E19f92CdBFF7e8FF0e2,0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54,r0,r1,500,10);
                        mstoreUniV3Pool(token0,token1,0x917933899c6a5F8E37F31E19f92CdBFF7e8FF0e2,0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54,r0,r1,3000,60);
                        mstoreUniV3Pool(token0,token1,0x917933899c6a5F8E37F31E19f92CdBFF7e8FF0e2,0xe34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54,r0,r1,10000,200);
                        mstoreUniV3Pool(token0,token1,0x91e1B99072f238352f59e58de875691e20Dc19c1,0x817e07951f93017a93327ac8cc31e946540203a19e1ecc37bc1761965c2d1090,r0,r1,100,1);
                        mstoreUniV3Pool(token0,token1,0x91e1B99072f238352f59e58de875691e20Dc19c1,0x817e07951f93017a93327ac8cc31e946540203a19e1ecc37bc1761965c2d1090,r0,r1,500,10);
                        mstoreUniV3Pool(token0,token1,0x91e1B99072f238352f59e58de875691e20Dc19c1,0x817e07951f93017a93327ac8cc31e946540203a19e1ecc37bc1761965c2d1090,r0,r1,3000,60);
                        mstoreUniV3Pool(token0,token1,0x91e1B99072f238352f59e58de875691e20Dc19c1,0x817e07951f93017a93327ac8cc31e946540203a19e1ecc37bc1761965c2d1090,r0,r1,10000,200);
                        mstoreAlgebraV3Pool(token0,token1,0x2D98E2FA9da15aa6dC9581AB097Ced7af697CB92,0x6ec6c9c8091d160c0aa74b2b14ba9c1717e95093bd3ac085cee99a49aab294a4,r0,r1);
                        uint len;
                        assembly{
                            len:=sub(sub(mload(0x40),_pools),0x20)
                        }
                        if(len>0){
                            assembly{
                                mstore(_pools,div(len,0x20))
                            }
                            pools[t1][t0]=pools[t0][t1]=_pools;
                        }
                    }
                }
            }
        }
    }

    function mstoreUniV2Pool(address t0,address t1, address factory, bytes32 poolInitCode,uint r0, uint r1) internal{
        unchecked{
            bytes32 fmp;
            assembly{fmp:=mload(0x40)}
            address pool=address(uint160(uint(keccak256(abi.encodePacked(hex'ff',factory, keccak256(abi.encodePacked(t0, t1)) ,poolInitCode)))));
            if(pool.code.length>0){
                uint reserve0; uint reserve1;
                assembly{
                    mstore(fmp,0x0902f1ac00000000000000000000000000000000000000000000000000000000)
                    pop(call(gas(), pool, 0, fmp, 0x04, fmp, 0x40))
                    reserve0:=mload(fmp)
                    reserve1:=mload(add(fmp,0x20))
                }
                if(reserve1>0 && reserve0>0&&(reserve1>(r1<<4) || reserve0>(r0<<4) || (reserve1*r0)/(reserve0+r0)>r1-(r1>>4) || (reserve0*r1)/(reserve1+r1)>r0-(r0>>4))){
                    assembly{
                        mstore(fmp,or(shl(128,reserve0),reserve1))
                        fmp:=add(fmp,0x20)
                        mstore(fmp,or(and(keccak256(fmp,0x20),B4_MASK),or(shl(216,1),or(shl(160,997000),and(pool,ADDR_MASK)))))
                        fmp:=add(fmp,0x20)
                    }
                }
            }
            assembly{mstore(0x40,fmp)}
        }
    }

    function mstoreUniV3Pool(address t0,address t1, address factory, bytes32 poolInitCode,uint r0, uint r1,uint fee,int s)internal {
        unchecked{
            bytes32 fmp;
            assembly{fmp:=mload(0x40)}
            address pool=address(uint160(uint(keccak256(abi.encodePacked(hex'ff',factory, keccak256(abi.encode(t0, t1,fee)),poolInitCode)))));
            if(pool.code.length>0){
                uint liquidity=IUniV3Pool(pool).liquidity();
                if(liquidity>2){                    
                    int t;uint sqrtPX64;
                    assembly{
                        mstore(fmp,0x3850c7bd00000000000000000000000000000000000000000000000000000000)
                        pop(call(gas(), pool, 0, fmp, 0x04, fmp, 0x40))
                        sqrtPX64 := shr(32,mload(fmp))
                        t:=mload(add(fmp,0x20))
                    }
                    (uint reserve0,uint reserve1,uint reserve0Limit,uint reserve1Limit)=reserves(liquidity,sqrtPX64,t,s);
                    if((r0+reserve0<reserve0Limit && (reserve0>(r0<<4) || (reserve1*r0)/(reserve0+r0)>r1-(r1>>4))) || (r1+reserve1<reserve1Limit && (reserve1>(r1<<4) || (reserve0*r1)/(reserve1+r1)>r0-(r0>>4)))){
                        assembly{
                            mstore(fmp,or(shl(128,reserve0),reserve1))
                            fmp:=add(fmp,0x20)
                            mstore(fmp,or(and(keccak256(fmp,0x20),B4_MASK),or(shl(160,sub(1000000,fee)),and(pool,ADDR_MASK))))
                            fmp:=add(fmp,0x20)
                            mstore(fmp,or(shl(128,reserve0Limit),reserve1Limit))
                            fmp:=add(fmp,0x20)
                        }
                    }
                }
            }
            assembly{mstore(0x40,fmp)}
        }
    }

    function mstoreAlgebraV3Pool(address t0,address t1, address factory, bytes32 poolInitCode,uint r0, uint r1)internal {
        unchecked{
            bytes32 fmp;
            assembly{fmp:=mload(0x40)}
            address pool =  address(uint160(uint(keccak256(abi.encodePacked(hex'ff',factory,keccak256(abi.encode(t0, t1)),poolInitCode)))));
            if(pool.code.length>0){
                uint liquidity =IAlgebraV3Pool(pool).liquidity();
                if(liquidity>2){
                    int t;uint sqrtPX64;
                    assembly{
                        mstore(fmp,0xe76c01e400000000000000000000000000000000000000000000000000000000)
                        pop(call(gas(), pool, 0, fmp, 0x04, fmp, 0x60))
                        sqrtPX64 := shr(32,mload(fmp))
                        t:=mload(add(fmp,0x20))
                    }
                    (uint reserve0,uint reserve1,uint reserve0Limit,uint reserve1Limit)=reserves(liquidity,sqrtPX64,t,60);
                    if((r0+reserve0<reserve0Limit && (reserve0>(r0<<4) || (reserve1*r0)/(reserve0+r0)>r1-(r1>>4))) || (r1+reserve1<reserve1Limit && (reserve1>(r1<<4) || (reserve0*r1)/(reserve1+r1)>r0-(r0>>4)))){
                        assembly{
                            let fee:=mload(add(fmp,0x40))
                            mstore(fmp,or(shl(128,reserve0),reserve1))
                            fmp:=add(fmp,0x20)
                            mstore(fmp,or(and(keccak256(fmp,0x20),B4_MASK),or(shl(216,2),or(shl(160,sub(1000000,fee)),and(pool,ADDR_MASK)))))
                            fmp:=add(fmp,0x20)
                            mstore(fmp,or(shl(128,reserve0Limit),reserve1Limit))
                            fmp:=add(fmp,0x20)
                        }
                    }
                }
            }
            assembly{mstore(0x40,fmp)}
        }
    }


    function reserves(uint liquidity,uint sqrtPX64,int t, int s)internal pure returns (uint reserve0,uint reserve1, uint reserve0Limit,uint reserve1Limit){
        unchecked{
            reserve0=(liquidity<<64)/sqrtPX64;
            reserve1=(liquidity*sqrtPX64)>>64;
            int tl;
            assembly {tl := mul(sub(sdiv(t, s), and(slt(t, 0), smod(t, s))), s)}
            int tu=tl+s;
            if(tl<MIN_TICK){
                tl=MIN_TICK;
            }
            else if(tu>MAX_TICK){
                tu=MAX_TICK;
            }
            reserve0Limit=(liquidity<<64)/(tSqrtPX64(tl)+1);
            reserve1Limit=((liquidity*tSqrtPX64(tu))>>64);
        }
    }

    function tSqrtPX64(int tick) internal pure returns (uint sqrtPriceX64) {
        unchecked {
            uint256 absTick;
            assembly {
                let mask := sar(255, tick)
                absTick := xor(mask, add(mask, tick))
            }
            uint256 price;
            assembly {
                price := xor(shl(128, 1), mul(xor(shl(128, 1), 0xfffcb933bd6fad37aa2d162d1a594001), and(absTick, 0x1)))
            }
            if (absTick & 0x2 != 0) price = (price * 0xfff97272373d413259a46990580e213a) >> 128;
            if (absTick & 0x4 != 0) price = (price * 0xfff2e50f5f656932ef12357cf3c7fdcc) >> 128;
            if (absTick & 0x8 != 0) price = (price * 0xffe5caca7e10e4e61c3624eaa0941cd0) >> 128;
            if (absTick & 0x10 != 0) price = (price * 0xffcb9843d60f6159c9db58835c926644) >> 128;
            if (absTick & 0x20 != 0) price = (price * 0xff973b41fa98c081472e6896dfb254c0) >> 128;
            if (absTick & 0x40 != 0) price = (price * 0xff2ea16466c96a3843ec78b326b52861) >> 128;
            if (absTick & 0x80 != 0) price = (price * 0xfe5dee046a99a2a811c461f1969c3053) >> 128;
            if (absTick & 0x100 != 0) price = (price * 0xfcbe86c7900a88aedcffc83b479aa3a4) >> 128;
            if (absTick & 0x200 != 0) price = (price * 0xf987a7253ac413176f2b074cf7815e54) >> 128;
            if (absTick & 0x400 != 0) price = (price * 0xf3392b0822b70005940c7a398e4b70f3) >> 128;
            if (absTick & 0x800 != 0) price = (price * 0xe7159475a2c29b7443b29c7fa6e889d9) >> 128;
            if (absTick & 0x1000 != 0) price = (price * 0xd097f3bdfd2022b8845ad8f792aa5825) >> 128;
            if (absTick & 0x2000 != 0) price = (price * 0xa9f746462d870fdf8a65dc1f90e061e5) >> 128;
            if (absTick & 0x4000 != 0) price = (price * 0x70d869a156d2a1b890bb3df62baf32f7) >> 128;
            if (absTick & 0x8000 != 0) price = (price * 0x31be135f97d08fd981231505542fcfa6) >> 128;
            if (absTick & 0x10000 != 0) price = (price * 0x9aa508b5b7a84e1c677de54f3e99bc9) >> 128;
            if (absTick & 0x20000 != 0) price = (price * 0x5d6af8dedb81196699c329225ee604) >> 128;
            if (absTick & 0x40000 != 0) price = (price * 0x2216e584f5fa1ea926041bedfe98) >> 128;
            if (absTick & 0x80000 != 0) price = (price * 0x48a170391f7dc42444e8fa2) >> 128;
            assembly {
                if sgt(tick, 0) { price := div(not(0), price) }
                sqrtPriceX64 := shr(64, price)
            }
        }
    }

}

//interfaces

interface IUniV2Pool{
    function getReserves()external view returns(uint reserve0, uint reserve1, uint blockTimestampLast);
}

interface IUniV3Pool{
    function slot0() external view returns(uint sqrtPX96, int t, uint observationIndex, uint observationCardinality, uint observationCardinalityNext, uint feeProtocol, bool unlocked);
    function liquidity() external view returns(uint liquidity);
}

interface IAlgebraV3Pool{
    function globalState() external view returns(uint sqrtPX96, int t, uint fee, uint timePointdex, uint comunityFeet0, uint comunityFeeT1, bool unlocked);
    function tickSpacing() external view returns(int s);
    function liquidity() external view returns(uint liquidity);
}


